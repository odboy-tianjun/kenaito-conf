package cn.odboy.confcenter.springboot;

import org.springframework.boot.autoconfigure.AutoConfiguration;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.context.annotation.Bean;
import org.springframework.core.env.Environment;

import cn.odboy.confcenter.ConfigcenterContext;
import cn.odboy.confcenter.RefreshScopeRegistry;

/**
 * Spring Boot 3 auto configuration for config center refresh support.
 *
 * <p>Registers a {@link RefreshScopeRegistry} bean post processor and captures the
 * {@link ApplicationContext} so that remote config changes can update the Spring
 * environment and refresh {@code @Value} / {@code @ConfigurationProperties} fields
 * at runtime.</p>
 *
 * <p>Registered through
 * {@code META-INF/spring/org.springframework.boot.autoconfigure.AutoConfiguration.imports}.</p>
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
@AutoConfiguration
public class KenaitoRefreshAutoConfiguration implements ApplicationContextAware {

	/**
	 * Must be static so the bean post processor is registered before regular beans
	 * are instantiated.
	 */
	@Bean
	public static RefreshScopeRegistry kenaitoRefreshScopeRegistry(Environment environment) {
		RefreshScopeRegistry registry = new RefreshScopeRegistry(environment);
		ConfigcenterContext.setRefreshScopeRegistry(registry);
		return registry;
	}

	@Override
	public void setApplicationContext(ApplicationContext applicationContext) {
		ConfigcenterContext.setApplicationContext(applicationContext);
	}
}
