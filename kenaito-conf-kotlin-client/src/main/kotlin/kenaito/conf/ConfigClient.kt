package kenaito.conf

import kotlinx.coroutines.*
import kotlinx.serialization.*
import kotlinx.serialization.json.*
import java.io.File
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.security.MessageDigest
import java.time.Duration
import java.util.concurrent.ConcurrentHashMap
import javax.crypto.Cipher
import javax.crypto.spec.IvParameterSpec
import javax.crypto.spec.SecretKeySpec

@Serializable
data class Config(
    val baseUrl: String = "",
    val appName: String = "",
    val env: String = "dev",
    val version: String = "latest",
    val token: String = "",
    val syncInterval: Int = 60,
    val ignoreGlobal: Boolean = false,
    val cacheDir: String = "."
)

object Crypto {
    private const val CRYPT_PREFIX = "{Cipher}"

    fun decryptConfig(secret: String, data: String): String {
        if (!data.startsWith(CRYPT_PREFIX)) return data
        return decryptWithAES(secret, data.removePrefix(CRYPT_PREFIX))
    }

    fun decryptWithAES(key: String, data: String): String {
        val md5 = MessageDigest.getInstance("MD5").digest(key.toByteArray()).joinToString("") { "%02x".format(it) }
        val aesKey = md5.substring(16).toByteArray()
        val cipherText = java.util.Base64.getDecoder().decode(data)
        val iv = IvParameterSpec(ByteArray(16))
        val cipher = Cipher.getInstance("AES/CBC/NoPadding")
        cipher.init(Cipher.DECRYPT_MODE, SecretKeySpec(aesKey, "AES"), iv)
        val plainText = cipher.doFinal(cipherText)
        val padding = plainText.last().toInt() and 0xFF
        return String(plainText, 0, plainText.size - padding)
    }
}

class Cache(private val cacheDir: String) {
    private val cacheFile = File(cacheDir, "config-cache.json")
    fun write(data: String) { cacheFile.writeText(data) }
    fun read(): String? = if (cacheFile.exists()) cacheFile.readText() else null
}

class ConfigClient(private val config: Config) {
    private val httpClient = HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(10)).build()
    private val configs = ConcurrentHashMap<String, String>()
    private var secret = ""
    private var globalSecret = ""
    private var lastTime = 0L
    private var started = false
    private val handlers = mutableListOf<(Map<String, String>) -> Unit>()
    private val json = Json { ignoreUnknownKeys = true }

    fun onChange(handler: (Map<String, String>) -> Unit) { handlers.add(handler) }
    fun get(key: String): String? = configs[key]
    fun getAll(): Map<String, String> = HashMap(configs)

    fun start() {
        if (started) return
        started = true
        try { fetchAll() } catch (e: Exception) {
            System.err.println("[ConfigClient] fetch all configs error: ${e.message}, trying cache")
            val cached = Cache(config.cacheDir).read()
            if (cached != null) {
                applyConfigs(cached)
                System.err.println("[ConfigClient] loaded ${configs.size} configs from cache")
            } else throw RuntimeException("fetch configs failed and no cache: ${e.message}")
        }
        CoroutineScope(Dispatchers.IO).launch {
            while (true) {
                delay(config.syncInterval * 1000L)
                try { fetchChanged() } catch (e: Exception) {
                    System.err.println("[ConfigClient] sync error: ${e.message}")
                }
            }
        }
    }

    fun stop() { started = false }

    private fun fetchAll() {
        val url = "${config.baseUrl}/api/fetchAllConfigs"
        val body = """{"appName":"${config.appName}","env":"${config.env}","version":"${config.version}","ignoreGlobal":${config.ignoreGlobal},"token":"${config.token}"}"""
        val req = HttpRequest.newBuilder().uri(URI.create(url)).POST(HttpRequest.BodyPublishers.ofString(body))
            .header("Content-Type", "application/json").timeout(Duration.ofSeconds(10)).build()
        val resp = httpClient.send(req, HttpResponse.BodyHandlers.ofString())
        if (resp.statusCode() != 200) throw RuntimeException("HTTP ${resp.statusCode()}")
        val data = json.parseToJsonElement(resp.body()).jsonObject
        if (data.containsKey("code") && data["code"]!!.jsonPrimitive.int != 0)
            throw RuntimeException("server error: ${data["code"]}")
        data["kenaito.config.encrypt-secret"]?.let { secret = it.jsonPrimitive.content }
        data["kenaito.config.global-encrypt-secret"]?.let { globalSecret = it.jsonPrimitive.content }
        data["latest.fetch.timestamp"]?.let { lastTime = it.jsonPrimitive.content.toLong() }
        Cache(config.cacheDir).write(resp.body())
        applyConfigs(resp.body())
        System.err.println("[ConfigClient] fetched ${configs.size} configs")
    }

    private fun fetchChanged() {
        val url = "${config.baseUrl}/api/fetchChangedConfigs"
        val body = """{"appName":"${config.appName}","env":"${config.env}","version":"${config.version}","lastTime":$lastTime,"ignoreGlobal":${config.ignoreGlobal},"token":"${config.token}"}"""
        val req = HttpRequest.newBuilder().uri(URI.create(url)).POST(HttpRequest.BodyPublishers.ofString(body))
            .header("Content-Type", "application/json").timeout(Duration.ofSeconds(10)).build()
        val resp = httpClient.send(req, HttpResponse.BodyHandlers.ofString())
        val result = json.parseToJsonElement(resp.body()).jsonObject
        if (!result.containsKey("code") || result["code"]!!.jsonPrimitive.int != 0)
            throw RuntimeException("server error")
        val changedData = result["data"]!!.jsonObject.toMutableMap()
        changedData.remove("latest.fetch.timestamp")?.let { lastTime = it.jsonPrimitive.content.toLong() }
        val changed = mutableMapOf<String, String>()
        changedData.forEach { (k, v) -> changed[k] = tryDecrypt(v.jsonPrimitive.content) }
        configs.putAll(changed)
        handlers.forEach { try { it(changed) } catch (e: Exception) { System.err.println("[ConfigClient] handler error: ${e.message}") } }
    }

    private fun applyConfigs(jsonStr: String) {
        val data = json.parseToJsonElement(jsonStr).jsonObject
        data.forEach { (k, v) ->
            if (!k.startsWith("kenaito.") && k != "latest.fetch.timestamp")
                configs[k] = tryDecrypt(v.jsonPrimitive.content)
        }
    }

    private fun tryDecrypt(value: String): String {
        var decrypted = Crypto.decryptConfig(secret, value)
        if (decrypted == value && globalSecret.isNotEmpty()) decrypted = Crypto.decryptConfig(globalSecret, value)
        return decrypted
    }
}
