#pragma once

#include <string>
#include <map>
#include <vector>
#include <functional>
#include <mutex>
#include <atomic>
#include <thread>
#include <chrono>

namespace kenaito {

struct Config {
    std::string base_url = "http://localhost:8080";
    std::string app_name;
    std::string env = "dev";
    std::string version = "latest";
    std::string token;
    int sync_interval = 60;
    bool ignore_global = false;
    std::string cache_dir = ".";
};

using ChangeHandler = std::function<void(const std::map<std::string, std::string>&)>;

class Cache {
public:
    explicit Cache(const std::string& dir);
    bool write(const std::string& json_str);
    std::string read();

private:
    std::string dir_;
    static const std::string CACHE_FILE;
};

class Listener {
public:
    void on_change(ChangeHandler handler);
    void notify(const std::map<std::string, std::string>& changed);

private:
    std::mutex mu_;
    std::vector<ChangeHandler> handlers_;
};

class Client {
public:
    explicit Client(const Config& config);
    ~Client();

    void on_change(ChangeHandler handler);
    std::string get(const std::string& key);
    std::map<std::string, std::string> get_all();
    bool start();
    void stop();
    void sync_once();

private:
    bool fetch_all();
    bool fetch_changed();
    void apply_configs(const std::map<std::string, std::string>& data);
    std::string http_post(const std::string& url, const std::string& body);

    Config config_;
    Cache cache_;
    Listener listener_;
    std::map<std::string, std::string> configs_;
    int64_t last_time_ = 0;
    std::mutex mu_;
    std::atomic<bool> started_{false};
    std::atomic<bool> stop_flag_{false};
    std::thread sync_thread_;
};

} // namespace kenaito
