import kenait.conf.Config
import kenait.conf.ConfigClient

fun main() {
    val configValues = mutableMapOf(
        "database.host" to (System.getenv("DATABASE_HOST") ?: "192.168.1.1"),
        "database.port" to (System.getenv("DATABASE_PORT") ?: "8888"),
        "database.name" to (System.getenv("DATABASE_NAME") ?: "xxxx"),
        "database.user" to (System.getenv("DATABASE_USER") ?: "xxxx"),
        "database.password" to (System.getenv("DATABASE_PASSWORD") ?: "xxxxxxxx"),
    )

    val config = Config(
        baseUrl = System.getenv("KENAITO_BASE_URL") ?: "http://localhost:8080",
        appName = System.getenv("KENAITO_APP_NAME") ?: "kenaito-springboot-demo",
        env = System.getenv("KENAITO_PROFILE") ?: "daily",
        version = System.getenv("KENAITO_VERSION") ?: "latest",
        token = System.getenv("KENAITO_TOKEN") ?: "",
        syncInterval = (System.getenv("KENAITO_SYNC_INTERVAL") ?: "10").toInt(),
        cacheDir = System.getenv("KENAITO_CACHE_DIR") ?: ".",
    )

    val client = ConfigClient(config)

    client.onChange { changed ->
        changed.forEach { (k, v) ->
            if (configValues.containsKey(k)) {
                configValues[k] = v
                System.err.println("config updated: $k = $v")
            }
        }
    }

    try {
        client.start()
        val remote = client.getAll()
        configValues.keys.forEach { k ->
            remote[k]?.let { configValues[k] = it }
        }
        System.err.println("loaded ${remote.size} configs from remote")
    } catch (e: Exception) {
        System.err.println("failed to init config client: ${e.message}")
    }

    println("Current configs:")
    configValues.forEach { (k, v) -> println("  $k = $v") }
    println("\nListening for config changes... (press Ctrl+C to stop)")
    Thread.currentThread().join()
}
