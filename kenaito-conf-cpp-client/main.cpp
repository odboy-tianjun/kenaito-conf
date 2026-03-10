#include <iostream>
#include <string>
#include <map>
#include <vector>
#include <functional>
#include <thread>
#include <mutex>
#include <atomic>
#include <chrono>
#include <fstream>
#include <sstream>
#include <cstring>
#include <curl/curl.h>

namespace kenaito_conf {

struct Config {
    std::string baseUrl;
    std::string appName;
    std::string env = "dev";
    std::string version = "latest";
    std::string token;
    int syncInterval = 60;
    bool ignoreGlobal = false;
    std::string cacheDir = ".";
};

class ConfigClient {
public:
    ConfigClient(const Config &config) : config_(config) {}

    void onChange(std::function<void(const std::map<std::string, std::string> &)> handler) {
        handlers_.push_back(handler);
    }

    std::string get(const std::string &key) {
        std::lock_guard<std::mutex> lock(mutex_);
        auto it = configs_.find(key);
        return it != configs_.end() ? it->second : "";
    }

    std::map<std::string, std::string> getAll() {
        std::lock_guard<std::mutex> lock(mutex_);
        return configs_;
    }

    void start() {
        if (started_) return;
        started_ = true;
        fetchAll();
        running_ = true;
        syncThread_ = std::thread([this]() {
            while (running_) {
                std::this_thread::sleep_for(std::chrono::seconds(config_.syncInterval));
                if (!running_) break;
                try { fetchChanged(); } catch (...) {}
            }
        });
    }

    void stop() {
        running_ = false;
        if (syncThread_.joinable()) syncThread_.join();
    }

private:
    Config config_;
    std::map<std::string, std::string> configs_;
    long lastTime_ = 0;
    bool started_ = false;
    std::atomic<bool> running_{false};
    std::mutex mutex_;
    std::thread syncThread_;
    std::vector<std::function<void(const std::map<std::string, std::string> &)>> handlers_;

    void fetchAll() { /* HTTP POST to /api/fetchAllConfigs */ }
    void fetchChanged() { /* HTTP POST to /api/fetchChangedConfigs */ }
};

} // namespace kenaito_conf

int main() {
    std::cout << "Kenaito Config Center C++ Client Demo" << std::endl;

    kenaito_conf::Config config;
    config.baseUrl = "http://localhost:8080";
    config.appName = "kenaito-springboot-demo";
    config.env = "daily";
    config.version = "latest";
    config.syncInterval = 30;

    kenaito_conf::ConfigClient client(config);

    client.onChange([](const std::map<std::string, std::string> &changed) {
        for (auto &[k, v] : changed)
            std::cout << "config changed: " << k << " = " << v << std::endl;
    });

    client.start();
    std::cout << "database.host = " << client.get("database.host") << std::endl;
    std::cout << "Listening for changes... (press Ctrl+C to stop)" << std::endl;

    std::cin.get();
    client.stop();
    return 0;
}
