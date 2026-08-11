#include "kenaito_conf/client.h"
#include <curl/curl.h>
#include <fstream>
#include <sstream>
#include <iostream>
#include <algorithm>

// nlohmann/json single header is bundled
#include "nlohmann/json.hpp"

using json = nlohmann::json;

namespace kenaito {

// ---- Cache ----

const std::string Cache::CACHE_FILE = "config-cache.json";

Cache::Cache(const std::string& dir) : dir_(dir) {}

bool Cache::write(const std::string& json_str) {
    std::string path = dir_ + "/" + CACHE_FILE;
    std::ofstream ofs(path);
    if (!ofs.is_open()) return false;
    ofs << json_str;
    return true;
}

std::string Cache::read() {
    std::string path = dir_ + "/" + CACHE_FILE;
    std::ifstream ifs(path);
    if (!ifs.is_open()) return "";
    std::stringstream ss;
    ss << ifs.rdbuf();
    return ss.str();
}

// ---- Listener ----

void Listener::on_change(ChangeHandler handler) {
    std::lock_guard<std::mutex> lock(mu_);
    handlers_.push_back(std::move(handler));
}

void Listener::notify(const std::map<std::string, std::string>& changed) {
    std::lock_guard<std::mutex> lock(mu_);
    for (auto& h : handlers_) {
        h(changed);
    }
}

// ---- Client ----

static size_t write_callback(void* contents, size_t size, size_t nmemb, void* userp) {
    auto* str = static_cast<std::string*>(userp);
    str->append(static_cast<char*>(contents), size * nmemb);
    return size * nmemb;
}

Client::Client(const Config& config)
    : config_(config), cache_(config.cache_dir) {
    if (config_.version.empty()) config_.version = "latest";
    if (config_.sync_interval <= 0) config_.sync_interval = 60;
    if (config_.cache_dir.empty()) config_.cache_dir = ".";
}

Client::~Client() {
    stop();
}

void Client::on_change(ChangeHandler handler) {
    listener_.on_change(std::move(handler));
}

std::string Client::get(const std::string& key) {
    std::lock_guard<std::mutex> lock(mu_);
    auto it = configs_.find(key);
    return it != configs_.end() ? it->second : "";
}

std::map<std::string, std::string> Client::get_all() {
    std::lock_guard<std::mutex> lock(mu_);
    return configs_;
}

bool Client::start() {
    if (started_.exchange(true)) return true;

    if (!fetch_all()) {
        std::cerr << "[ConfigClient] fetch all configs error, trying cache" << std::endl;
        std::string cached = cache_.read();
        if (!cached.empty()) {
            try {
                auto j = json::parse(cached);
                std::map<std::string, std::string> data;
                for (auto& [k, v] : j.items()) {
                    if (v.is_string()) data[k] = v.get<std::string>();
                }
                apply_configs(data);
                std::cerr << "[ConfigClient] loaded " << configs_.size() << " configs from cache" << std::endl;
            } catch (const std::exception& e) {
                std::cerr << "[ConfigClient] cache parse error: " << e.what() << std::endl;
                return false;
            }
        } else {
            std::cerr << "[ConfigClient] fetch configs failed and no cache" << std::endl;
            return false;
        }
    }

    return true;
}

void Client::stop() {
    stop_flag_ = true;
    if (sync_thread_.joinable()) {
        sync_thread_.join();
    }
}

void Client::sync_once() {
    if (!fetch_changed()) {
        std::cerr << "[ConfigClient] sync error" << std::endl;
    }
}

std::string Client::http_post(const std::string& url, const std::string& body) {
    CURL* curl = curl_easy_init();
    if (!curl) return "";

    std::string response;
    struct curl_slist* headers = nullptr;
    headers = curl_slist_append(headers, "Content-Type: application/json");

    curl_easy_setopt(curl, CURLOPT_URL, url.c_str());
    curl_easy_setopt(curl, CURLOPT_POSTFIELDS, body.c_str());
    curl_easy_setopt(curl, CURLOPT_HTTPHEADER, headers);
    curl_easy_setopt(curl, CURLOPT_WRITEFUNCTION, write_callback);
    curl_easy_setopt(curl, CURLOPT_WRITEDATA, &response);
    curl_easy_setopt(curl, CURLOPT_TIMEOUT, 10L);

    CURLcode res = curl_easy_perform(curl);
    curl_slist_free_all(headers);
    curl_easy_cleanup(curl);

    if (res != CURLE_OK) return "";
    return response;
}

bool Client::fetch_all() {
    std::string url = config_.base_url + "/api/fetchAllConfigs";

    json payload;
    payload["appName"] = config_.app_name;
    payload["env"] = config_.env;
    payload["version"] = config_.version;
    payload["ignoreGlobal"] = config_.ignore_global;
    payload["token"] = config_.token;

    std::string resp = http_post(url, payload.dump());
    if (resp.empty()) return false;

    try {
        auto j = json::parse(resp);

        if (j.contains("code") && j["code"].get<int>() != 0) {
            std::string msg = j.value("msg", "unknown");
            std::cerr << "[ConfigClient] server error: " << msg << std::endl;
            return false;
        }

        if (j.contains("latest.fetch.timestamp")) {
            last_time_ = std::stoll(j["latest.fetch.timestamp"].get<std::string>());
        }

        // cache raw response
        cache_.write(resp);

        // apply configs
        std::map<std::string, std::string> data;
        for (auto& [k, v] : j.items()) {
            if (v.is_string()) data[k] = v.get<std::string>();
        }
        apply_configs(data);

        std::cerr << "[ConfigClient] fetched " << configs_.size() << " configs" << std::endl;
        return true;
    } catch (const std::exception& e) {
        std::cerr << "[ConfigClient] parse error: " << e.what() << std::endl;
        return false;
    }
}

bool Client::fetch_changed() {
    std::string url = config_.base_url + "/api/fetchChangedConfigs";

    json payload;
    payload["appName"] = config_.app_name;
    payload["env"] = config_.env;
    payload["version"] = config_.version;
    payload["lastTime"] = last_time_;
    payload["ignoreGlobal"] = config_.ignore_global;

    std::string resp = http_post(url, payload.dump());
    if (resp.empty()) {
        std::cerr << "[ConfigClient] fetch_changed: empty response (network error or timeout)" << std::endl;
        return false;
    }

    try {
        auto j = json::parse(resp);

        if (j.contains("code") && j["code"].get<int>() != 0) {
            std::string msg = j.value("msg", "unknown");
            std::cerr << "[ConfigClient] fetch_changed: server error code=" << j["code"] << " msg=" << msg << std::endl;
            return false;
        }

        if (!j.contains("data")) {
            std::cerr << "[ConfigClient] fetch_changed: missing 'data' field in response" << std::endl;
            return false;
        }

        auto data = j["data"];

        if (data.contains("latest.fetch.timestamp")) {
            last_time_ = std::stoll(data["latest.fetch.timestamp"].get<std::string>());
            data.erase("latest.fetch.timestamp");
        }

        // handle deleted keys
        if (data.contains("deletedKeys") && !data["deletedKeys"].get<std::string>().empty()) {
            std::string deleted_str = data["deletedKeys"].get<std::string>();
            std::istringstream iss(deleted_str);
            std::string key;
            std::vector<std::string> deleted_keys;
            while (std::getline(iss, key, ',')) {
                std::lock_guard<std::mutex> lock(mu_);
                configs_.erase(key);
                deleted_keys.push_back(key);
            }
            
            std::string keys;
            for (const auto& k : deleted_keys) {
                if (!keys.empty()) keys += ", ";
                keys += k;
            }
            std::cerr << "[ConfigClient] " << deleted_keys.size() << " configs deleted: " << keys << std::endl;
            data.erase("deletedKeys");
        }

        // handle changed configs
        std::map<std::string, std::string> changed;
        for (auto& [k, v] : data.items()) {
            if (v.is_string()) {
                changed[k] = v.get<std::string>();
            }
        }

        if (!changed.empty()) {
            {
                std::lock_guard<std::mutex> lock(mu_);
                for (auto& [k, v] : changed) {
                    configs_[k] = v;
                }
            }
            listener_.notify(changed);

            std::string keys;
            for (auto& [k, v] : changed) {
                if (!keys.empty()) keys += ", ";
                keys += k;
            }
            std::cerr << "[ConfigClient] " << changed.size() << " configs changed: " << keys << std::endl;
        }

        return true;
    } catch (const std::exception& e) {
        std::cerr << "[ConfigClient] fetch_changed: parse error: " << e.what() << std::endl;
        std::cerr << "[ConfigClient] response was: " << resp.substr(0, 200) << std::endl;
        return false;
    }
}

void Client::apply_configs(const std::map<std::string, std::string>& data) {
    std::lock_guard<std::mutex> lock(mu_);
    for (auto& [k, v] : data) {
        if (k.substr(0, 8) == "kenaito." || k == "latest.fetch.timestamp") continue;
        configs_[k] = v;
    }
}

} // namespace kenaito
