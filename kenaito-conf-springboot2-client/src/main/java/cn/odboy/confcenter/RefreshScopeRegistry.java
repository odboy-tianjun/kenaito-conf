package cn.odboy.confcenter;

import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.beans.factory.config.BeanPostProcessor;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.core.env.Environment;

/**
 * Refresh Scope Registry
 * Supports dynamic refresh of @Value and @ConfigurationProperties fields
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-16
 */
public class RefreshScopeRegistry implements BeanPostProcessor {

	private static final Logger logger = LoggerFactory.getLogger(RefreshScopeRegistry.class);

	private Environment environment;
	private final List<ValueFieldHolder> valueFields = new ArrayList<>();
	private final List<ConfigPropsHolder> configPropsBeans = new ArrayList<>();

	public RefreshScopeRegistry(Environment environment) {
		this.environment = environment;
	}

	public int getValueFieldCount() {
		return valueFields.size();
	}

	public int getConfigPropsCount() {
		return configPropsBeans.size();
	}

	@Override
	public Object postProcessAfterInitialization(Object bean, String beanName) throws BeansException {
		Class<?> clazz = bean.getClass();

		// Record @Value fields
		for (Field field : clazz.getDeclaredFields()) {
			if (field.isAnnotationPresent(Value.class)) {
				valueFields.add(new ValueFieldHolder(bean, field));
			}
		}

		// Record @ConfigurationProperties beans
		if (clazz.isAnnotationPresent(ConfigurationProperties.class)) {
			configPropsBeans.add(new ConfigPropsHolder(bean));
		}

		return bean;
	}

	/**
	 * Refresh all @Value fields and @ConfigurationProperties beans
	 */
	public void refresh() {
		refreshValueFields();
		refreshConfigPropsBeans();
	}

	private void refreshValueFields() {
		for (ValueFieldHolder holder : valueFields) {
			try {
				holder.refresh(environment);
			} catch (Exception e) {
				// ignore refresh errors
			}
		}
	}

	private void refreshConfigPropsBeans() {
		for (ConfigPropsHolder holder : configPropsBeans) {
			try {
				holder.refresh(environment);
			} catch (Exception e) {
				// ignore refresh errors
			}
		}
	}

	/**
	 * Holder for @Value field
	 */
	private static class ValueFieldHolder {
		private final Object bean;
		private final Field field;
		private final String key;

		ValueFieldHolder(Object bean, Field field) {
			this.bean = bean;
			this.field = field;
			this.key = parseKey(field.getAnnotation(Value.class).value());
		}

		private String parseKey(String valueAnnotation) {
			// Extract key from "${database.host:default}" -> "database.host"
			String cleaned = valueAnnotation.replace("${", "").replace("}", "");
			int colonIdx = cleaned.indexOf(':');
			return colonIdx > 0 ? cleaned.substring(0, colonIdx) : cleaned;
		}

		void refresh(Environment env) throws Exception {
			String newValue = env.getProperty(key);
			if (newValue != null) {
				field.setAccessible(true);
				Object oldValue = field.get(bean);
				Object convertedValue = convertType(field.getType(), newValue);
				field.set(bean, convertedValue);
				
				// 验证是否设置成功
				Object verifyValue = field.get(bean);
				boolean success = (convertedValue != null && convertedValue.equals(verifyValue));
				logger.info("[RefreshScope] Refreshed @Value field: {}.{} | old='{}' | new='{}' | converted='{}' | verify='{}' | success={}", 
					bean.getClass().getName(), field.getName(), oldValue, newValue, convertedValue, verifyValue, success);
				
				if (!success) {
					logger.warn("[RefreshScope] Field refresh failed! bean={}, field={}, fieldType={}, newValueClass={}", 
						bean.getClass().getName(), field.getName(), field.getType(), convertedValue != null ? convertedValue.getClass() : "null");
				}
			}
		}

		private Object convertType(Class<?> type, String value) {
			if (type == String.class) return value;
			if (type == int.class || type == Integer.class) return Integer.parseInt(value);
			if (type == long.class || type == Long.class) return Long.parseLong(value);
			if (type == boolean.class || type == Boolean.class) return Boolean.parseBoolean(value);
			if (type == double.class || type == Double.class) return Double.parseDouble(value);
			return value;
		}
	}

	/**
	 * Holder for @ConfigurationProperties bean
	 */
	private static class ConfigPropsHolder {
		private final Object bean;

		ConfigPropsHolder(Object bean) {
			this.bean = bean;
		}

		void refresh(Environment env) {
			Class<?> clazz = bean.getClass();
			ConfigurationProperties anno = clazz.getAnnotation(ConfigurationProperties.class);
			if (anno != null) {
				String prefix = anno.prefix();
				bindProperties(bean, prefix, env);
			}
		}

		private void bindProperties(Object bean, String prefix, Environment env) {
			try {
				Class<?> clazz = bean.getClass();
				for (Field field : getAllDeclaredFields(clazz)) {
					String propertyName = prefix + "." + camelToKebab(field.getName());
					String value = env.getProperty(propertyName);
					if (value != null) {
						field.setAccessible(true);
						field.set(bean, convertType(field.getType(), value));
					}
				}
			} catch (Exception e) {
				// ignore
			}
		}

		private Field[] getAllDeclaredFields(Class<?> clazz) {
			List<Field> fields = new ArrayList<>();
			while (clazz != null && clazz != Object.class) {
				for (Field field : clazz.getDeclaredFields()) {
					fields.add(field);
				}
				clazz = clazz.getSuperclass();
			}
			return fields.toArray(new Field[0]);
		}

		private String camelToKebab(String camelCase) {
			StringBuilder result = new StringBuilder();
			for (int i = 0; i < camelCase.length(); i++) {
				char c = camelCase.charAt(i);
				if (Character.isUpperCase(c)) {
					if (i > 0) result.append('-');
					result.append(Character.toLowerCase(c));
				} else {
					result.append(c);
				}
			}
			return result.toString();
		}

		private Object convertType(Class<?> type, String value) {
			if (type == String.class) return value;
			if (type == int.class || type == Integer.class) return Integer.parseInt(value);
			if (type == long.class || type == Long.class) return Long.parseLong(value);
			if (type == boolean.class || type == Boolean.class) return Boolean.parseBoolean(value);
			if (type == double.class || type == Double.class) return Double.parseDouble(value);
			return value;
		}
	}
}
