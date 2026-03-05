plugins {
    kotlin("jvm") version "1.9.0"
    application
}

repositories { mavenCentral() }

dependencies {
    implementation(files("../kenaito-conf-kotlin-client/build/libs/kenaito-conf-kotlin-client-1.0.0.jar"))
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.7.3")
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.0")
}

application { mainClass.set("demo.AppKt") }
