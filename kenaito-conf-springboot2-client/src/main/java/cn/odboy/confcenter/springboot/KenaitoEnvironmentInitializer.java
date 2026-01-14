package cn.odboy.confcenter.springboot;

import java.util.Properties;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.env.EnvironmentPostProcessor;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.PropertiesPropertySource;
import org.springframework.core.env.StandardEnvironment;

import cn.odboy.confcenter.ConfigcenterContext;

/**
 * Spring Boot 2 Environment Post Processor
 * Initializes config center before application context is created
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-11
 */
public class KenaitoEnvironmentInitializer implements EnvironmentPostProcessor {

	private ConfigcenterContext ccContext = ConfigcenterContext.getInstance();

	@Override
	public void postProcessEnvironment(ConfigurableEnvironment environment, SpringApplication application) {
		if (ccContext.isRemoteEnabled() && !ccContext.isProcessed()) {
			Properties properties = new Properties();
			ccContext.init(properties, true);
			ccContext.mergeRemoteProperties(properties);
			PropertiesPropertySource propertySource = new PropertiesPropertySource(ConfigcenterContext.MANAGER_PROPERTY_SOURCE, properties);
			environment.getPropertySources().addFirst(propertySource);
		}
	}
}
