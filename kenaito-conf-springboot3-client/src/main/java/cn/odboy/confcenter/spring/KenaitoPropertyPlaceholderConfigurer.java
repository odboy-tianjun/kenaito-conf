package cn.odboy.confcenter.spring;

import java.io.IOException;
import java.util.Properties;

import org.springframework.beans.BeansException;
import org.springframework.beans.factory.DisposableBean;
import org.springframework.context.ApplicationContext;
import org.springframework.context.ApplicationContextAware;
import org.springframework.context.support.PropertySourcesPlaceholderConfigurer;

import cn.odboy.confcenter.ConfigcenterContext;

/**
 * Spring property placeholder configurer for plain Spring (non-Boot) applications.
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
public class KenaitoPropertyPlaceholderConfigurer extends PropertySourcesPlaceholderConfigurer
		implements DisposableBean, ApplicationContextAware {

	private boolean remoteEnabledSet = false;

	private final ConfigcenterContext ccContext = ConfigcenterContext.getInstance();

	public void setRemoteEnabled(boolean remoteEnabled) {
		ccContext.setRemoteEnabled(remoteEnabled);
		remoteEnabledSet = true;
	}

	@Override
	protected Properties mergeProperties() throws IOException {
		Properties properties = super.mergeProperties();
		if (!remoteEnabledSet) {
			ccContext.setRemoteEnabled(Boolean.parseBoolean(
					properties.getProperty("kenaito.config.enabled",
							properties.getProperty("kenaito.configcenter.enabled", "true"))));
		}
		ccContext.init(properties, false);
		ccContext.mergeRemoteProperties(properties);
		return properties;
	}

	@Override
	public void destroy() {
		ccContext.close();
	}

	@Override
	public void setApplicationContext(ApplicationContext applicationContext) throws BeansException {
		ConfigcenterContext.setApplicationContext(applicationContext);
	}
}
