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
#include <openssl/evp.h>
#include <openssl/md5.h>

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

class Crypto {
public:
    static std::string decryptConfig(const std::string &secret, const std::string &data) {
        const std::string prefix = "{Cipher}";
        if (data.substr(0, prefix.size()) != prefix) return data;
        return decryptWithAES(secret, data.substr(prefix.size()));
    }

    static std::string decryptWithAES(const std::string &key, const std::string &data) {
        unsigned char digest[MD5_DIGEST_LENGTH];
        MD5((unsigned char *)key.c_str(), key.size(), digest);
        char md5Hex[33];
        for (int i = 0; i < MD5_DIGEST_LENGTH; i++)
            sprintf(md5Hex + i * 2, "%02x", digest[i]);
        md5Hex[32] = 0;

        std::string aesKey(md5Hex + 16, 16);

        // Base64 decode
        BIO *b64 = BIO_new(BIO_f_base64());
        BIO *mem = BIO_new_mem_buf(data.c_str(), data.size());
        mem = BIO_push(b64, mem);
        BIO_set_flags(mem, BIO_FLAGS_BASE64_NO_NL);
        std::vector<unsigned char> cipherText(data.size());
        int cipherLen = BIO_read(mem, cipherText.data(), data.size());
        BIO_free_all(mem);

        unsigned char iv[16] = {0};
        std::vector<unsigned char> plainText(cipherLen + 16);
        EVP_CIPHER_CTX *ctx = EVP_CIPHER_CTX_new();
        EVP_DecryptInit_ex(ctx, EVP_aes_128_cbc(), nullptr, (unsigned char *)aesKey.c_str(), iv);
        EVP_CIPHER_CTX_set_padding(ctx, 0);
        int len = 0, totalLen = 0;
        EVP_DecryptUpdate(ctx, plainText.data(), &len, cipherText.data(), cipherLen);
        totalLen = len;
        EVP_DecryptFinal_ex(ctx, plainText.data() + len, &len);
        totalLen += len;
        EVP_CIPHER_CTX_free(ctx);

        int padding = plainText[totalLen - 1];
        return std::string(plainText.begin(), plainText.begin() + totalLen - padding);
    }
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
    std::string secret_, globalSecret_;
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
