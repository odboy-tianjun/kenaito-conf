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
 * Spring Property Placeholder Configurer for non-SpringBoot applications
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-11
 */
public class KenaitoPropertyPlaceholderConfigurer extends PropertySourcesPlaceholderConfigurer
		implements DisposableBean, ApplicationContextAware {

	private boolean setRemoteEnabled = false;
	private ConfigcenterContext ccContext = ConfigcenterContext.getInstance();

	public void setRemoteEnabled(boolean remoteEnabled) {
		ccContext.setRemoteEnabled(remoteEnabled);
		setRemoteEnabled = true;
	}

	@Override
	protected Properties mergeProperties() throws IOException {
		Properties properties = super.mergeProperties();
		if (!setRemoteEnabled) {
			ccContext.setRemoteEnabled(Boolean.parseBoolean(properties.getProperty("kenaito.configcenter.enabled", "true")));
		}
		ccContext.init(properties, false);
		ccContext.mergeRemoteProperties(properties);
		return properties;
	}

	@Override
	public void destroy() throws Exception {
		ccContext.close();
	}

	@Override
	public void setApplicationContext(ApplicationContext applicationContext) throws BeansException {
		ConfigcenterContext.setApplicationContext(applicationContext);
	}
}
