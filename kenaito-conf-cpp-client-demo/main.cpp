#include <kenaito_conf/client.h>
#include <iostream>
#include <cstdlib>
#include <csignal>
#include <thread>
#include <chrono>
#include <map>

static kenaito::Client* g_client = nullptr;

static void signal_handler(int) {
    std::cout << "\nShutting down..." << std::endl;
    if (g_client) g_client->stop();
}

static std::string get_env(const char* key, const std::string& default_val) {
    const char* val = std::getenv(key);
    return (val && val[0] != '\0') ? val : default_val;
}

static int get_env_int(const char* key, int default_val) {
    const char* val = std::getenv(key);
    if (val && val[0] != '\0') {
        try { return std::stoi(val); } catch (...) {}
    }
    return default_val;
}

int main() {
    std::map<std::string, std::string> config_values = {
        {"database.host", "192.168.1.1"},
        {"database.port", "8888"},
        {"database.name", "xxxx"},
        {"database.user", "xxxx"},
        {"database.password", "xxxxxxxx"},
    };

    kenaito::Config config;
    config.base_url = get_env("KENAITO_BASE_URL", "http://localhost:8080");
    config.app_name = get_env("KENAITO_APP_NAME", "kenaito-springboot-demo");
    config.env = get_env("KENAITO_ENV", "daily");
    config.version = get_env("KENAITO_VERSION", "latest");
    config.token = get_env("KENAITO_TOKEN", "");
    config.sync_interval = get_env_int("KENAITO_SYNC_INTERVAL", 10);
    config.cache_dir = get_env("KENAITO_CACHE_DIR", ".");

    kenaito::Client client(config);
    g_client = &client;

    client.on_change([&config_values](const std::map<std::string, std::string>& changed) {
        std::cout << "Config changed:" << std::endl;
        for (const auto& kv : changed) {
            std::cout << "  " << kv.first << " = " << kv.second << std::endl;
            auto it = config_values.find(kv.first);
            if (it != config_values.end()) {
                it->second = kv.second;
                std::cout << "  [updated] " << kv.first << " = " << kv.second << std::endl;
            }
        }
    });

    if (!client.start()) {
        std::cerr << "Failed to init config client, using local defaults" << std::endl;
    }

    auto remote = client.get_all();
    for (auto& kv : config_values) {
        auto it = remote.find(kv.first);
        if (it != remote.end()) {
            kv.second = it->second;
        }
    }
    std::cout << "loaded " << remote.size() << " configs from remote" << std::endl;

    std::cout << "\nCurrent configs:" << std::endl;
    for (const auto& kv : config_values) {
        std::cout << "  " << kv.first << " = " << kv.second << std::endl;
    }

    std::cout << "\nListening for config changes... (press Ctrl+C to stop)" << std::endl;

    std::signal(SIGINT, signal_handler);

    while (true) {
        client.sync_once();
        std::this_thread::sleep_for(std::chrono::seconds(config.sync_interval));
    }

    return 0;
}
