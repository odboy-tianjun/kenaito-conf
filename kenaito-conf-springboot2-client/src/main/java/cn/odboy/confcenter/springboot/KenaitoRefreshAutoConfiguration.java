package cn.odboy.confcenter.springboot;

import org.springframework.beans.BeansException;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.context.annotation.Configuration;

import cn.odboy.confcenter.ConfigcenterContext;

/**
 * Spring Boot auto configuration for config center refresh support.
 *
 * Captures the Spring ApplicationContext into {@link ConfigcenterContext} so that
 * remote config changes (detected by the polling listener) can update the Spring
 * Environment property source and refresh {@code @Value} / {@code @ConfigurationProperties}
 * fields at runtime.
 *
 * The non-SpringBoot path sets the context via CCPropertyPlaceholderConfigurer;
 * this class provides the equivalent wiring for Spring Boot applications.
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-16
 */
@Configuration
public class KenaitoRefreshAutoConfiguration implements ApplicationContextAware {

	@Override
	public void setApplicationContext(ApplicationContext applicationContext) throws BeansException {
		ConfigcenterContext.setApplicationContext(applicationContext);
	}
}
