package cn.odboy.confcenter.springboot;

import java.util.Properties;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.env.EnvironmentPostProcessor;
import org.springframework.core.Ordered;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.EnumerablePropertySource;
import org.springframework.core.env.PropertiesPropertySource;

import cn.odboy.confcenter.ConfigcenterContext;

/**
 * Spring Boot environment post processor.
 *
 * <p>Initializes the config center before the application context is created and
 * injects the remote configuration as a high-priority property source.</p>
 *
 * <p>Registered through {@code META-INF/spring.factories}, which Spring Boot 3
 * still supports for {@link EnvironmentPostProcessor} implementations.</p>
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
public class KenaitoEnvironmentInitializer implements EnvironmentPostProcessor, Ordered {

	private final ConfigcenterContext ccContext = ConfigcenterContext.getInstance();

	@Override
	public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
		if (!ccContext.isRemoteEnabled() || ccContext.isProcessed()) {
			return;
		}
		Properties properties = new Properties();
		// Extract config center settings from the Spring Environment. The config data
		// post processor runs earlier, so application.properties is already loaded.
		environment.getPropertySources().forEach(propertySource -> {
			if (propertySource instanceof EnumerablePropertySource<?> enumerable) {
				for (String key : enumerable.getPropertyNames()) {
					if (key.startsWith("kenaito.config.")
							|| key.startsWith("kenaito.configcenter.")
							|| key.equals("spring.application.name")) {
						String value = environment.getProperty(key);
						if (value != null) {
							properties.setProperty(key, value);
						}
					}
				}
			}
		});
		ccContext.init(properties, true);
		ccContext.mergeRemoteProperties(properties);
		PropertiesPropertySource propertySource =
				new PropertiesPropertySource(ConfigcenterContext.MANAGER_PROPERTY_SOURCE, properties);
		environment.getPropertySources().addFirst(propertySource);
	}

	@Override
	public int getOrder() {
		// Run after ConfigDataEnvironmentPostProcessor (HIGHEST_PRECEDENCE + 10).
		return Ordered.LOWEST_PRECEDENCE;
	}
}
