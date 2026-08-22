package cn.odboy.confcenter;

import java.lang.reflect.Field;
import java.lang.reflect.Modifier;
import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Set;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.beans.BeansException;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.beans.factory.config.BeanPostProcessor;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.core.annotation.AnnotatedElementUtils;
import org.springframework.core.env.Environment;
import org.springframework.util.ClassUtils;

/**
 * Refresh scope registry.
 *
 * <p>Collects {@code @Value} fields and {@code @ConfigurationProperties} beans while
 * the application context starts, then refreshes only the fields affected by a
 * config change. This is registered as a static bean by
 * {@link cn.odboy.confcenter.springboot.KenaitoRefreshAutoConfiguration}.</p>
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
public class RefreshScopeRegistry implements BeanPostProcessor {

	private static final Logger logger = LoggerFactory.getLogger(RefreshScopeRegistry.class);

	private final Environment environment;

	private final Map<String, ValueFieldHolder> valueFields = new LinkedHashMap<>();

	private final Map<String, ConfigPropsHolder> configPropsBeans = new LinkedHashMap<>();

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
		if (bean == null) {
			return null;
		}
		Class<?> clazz = ClassUtils.getUserClass(bean);

		for (Field field : getAllDeclaredFields(clazz)) {
			if (Modifier.isStatic(field.getModifiers())) {
				continue;
			}
			Value valueAnnotation = field.getAnnotation(Value.class);
			if (valueAnnotation == null) {
				continue;
			}
			String key = parseKey(valueAnnotation.value());
			if (key == null) {
				continue;
			}
			valueFields.putIfAbsent(beanName + "#" + field.getName(), new ValueFieldHolder(bean, field, key));
		}

		ConfigurationProperties annotation =
				AnnotatedElementUtils.findMergedAnnotation(clazz, ConfigurationProperties.class);
		if (annotation != null) {
			String prefix = firstNonBlank(annotation.prefix(), annotation.value());
			if (prefix != null) {
				configPropsBeans.putIfAbsent(beanName, new ConfigPropsHolder(bean, prefix));
			}
		}
		return bean;
	}

	/**
	 * Refreshes the fields affected by the supplied keys.
	 *
	 * @return number of refreshed {@code @Value} fields
	 */
	public synchronized int refresh(Set<String> changedKeys) {
		int count = 0;
		for (ValueFieldHolder holder : valueFields.values()) {
			if (changedKeys.contains(holder.key)) {
				if (holder.refresh(environment)) {
					count++;
				}
			}
		}
		for (ConfigPropsHolder holder : configPropsBeans.values()) {
			try {
				holder.refresh(environment, changedKeys);
			} catch (Exception e) {
				logger.warn("[RefreshScope] refresh @ConfigurationProperties bean {} error: {}",
						holder.bean.getClass().getName(), e.getMessage());
			}
		}
		return count;
	}

	private static String parseKey(String expression) {
		if (expression == null) {
			return null;
		}
		int start = expression.indexOf("${");
		if (start < 0) {
			return null;
		}
		int end = expression.indexOf('}', start);
		if (end < 0) {
			return null;
		}
		String key = expression.substring(start + 2, end);
		int colon = key.indexOf(':');
		if (colon >= 0) {
			key = key.substring(0, colon);
		}
		key = key.trim();
		return key.isEmpty() ? null : key;
	}

	private static String firstNonBlank(String... values) {
		for (String value : values) {
			if (value != null && !value.isBlank()) {
				return value;
			}
		}
		return null;
	}

	private static List<Field> getAllDeclaredFields(Class<?> clazz) {
		List<Field> fields = new ArrayList<>();
		Class<?> current = clazz;
		while (current != null && current != Object.class) {
			for (Field field : current.getDeclaredFields()) {
				fields.add(field);
			}
			current = current.getSuperclass();
		}
		return fields;
	}

	private static String camelToKebab(String camelCase) {
		StringBuilder result = new StringBuilder();
		for (int i = 0; i < camelCase.length(); i++) {
			char c = camelCase.charAt(i);
			if (Character.isUpperCase(c)) {
				if (i > 0) {
					result.append('-');
				}
				result.append(Character.toLowerCase(c));
			} else {
				result.append(c);
			}
		}
		return result.toString();
	}

	private static Object convertType(Class<?> type, String value) {
		if (type == String.class || type == Object.class) {
			return value;
		}
		if (type == int.class || type == Integer.class) {
			return Integer.parseInt(value.trim());
		}
		if (type == long.class || type == Long.class) {
			return Long.parseLong(value.trim());
		}
		if (type == boolean.class || type == Boolean.class) {
			return Boolean.parseBoolean(value.trim());
		}
		if (type == double.class || type == Double.class) {
			return Double.parseDouble(value.trim());
		}
		if (type == float.class || type == Float.class) {
			return Float.parseFloat(value.trim());
		}
		if (type == short.class || type == Short.class) {
			return Short.parseShort(value.trim());
		}
		if (type == char.class || type == Character.class) {
			return value.isEmpty() ? null : value.charAt(0);
		}
		if (type == byte.class || type == Byte.class) {
			return Byte.parseByte(value.trim());
		}
		return value;
	}

	/**
	 * Holder for a {@code @Value} annotated field.
	 */
	private static class ValueFieldHolder {

		private final Object bean;

		private final Field field;

		private final String key;

		ValueFieldHolder(Object bean, Field field, String key) {
			this.bean = bean;
			this.field = field;
			this.key = key;
		}

		boolean refresh(Environment env) {
			String newValue = env.getProperty(key);
			if (newValue == null) {
				return false;
			}
			try {
				field.setAccessible(true);
				Object oldValue = field.get(bean);
				Object convertedValue = convertType(field.getType(), newValue);
				field.set(bean, convertedValue);
				logger.info("[RefreshScope] refreshed @Value field: {}.{} | old='{}' | new='{}'",
						bean.getClass().getName(), field.getName(), oldValue, newValue);
				return true;
			} catch (Exception e) {
				logger.warn("[RefreshScope] refresh @Value field {}.{} error: {}",
						bean.getClass().getName(), field.getName(), e.getMessage());
				return false;
			}
		}
	}

	/**
	 * Holder for a {@code @ConfigurationProperties} bean.
	 */
	private static class ConfigPropsHolder {

		private final Object bean;

		private final String prefix;

		ConfigPropsHolder(Object bean, String prefix) {
			this.bean = bean;
			this.prefix = prefix;
		}

		void refresh(Environment env, Set<String> changedKeys) {
			for (Field field : getAllDeclaredFields(ClassUtils.getUserClass(bean))) {
				if (Modifier.isStatic(field.getModifiers()) || Modifier.isFinal(field.getModifiers())) {
					continue;
				}
				String kebabKey = prefix + "." + camelToKebab(field.getName());
				String plainKey = prefix + "." + field.getName();
				String matchedKey = changedKeys.contains(kebabKey) ? kebabKey
						: changedKeys.contains(plainKey) ? plainKey : null;
				if (matchedKey == null) {
					continue;
				}
				String value = env.getProperty(matchedKey);
				if (value == null) {
					continue;
				}
				try {
					field.setAccessible(true);
					field.set(bean, convertType(field.getType(), value));
					logger.info("[RefreshScope] refreshed @ConfigurationProperties field: {}.{} = '{}'",
							bean.getClass().getName(), field.getName(), value);
				} catch (Exception e) {
					logger.warn("[RefreshScope] refresh @ConfigurationProperties field {}.{} error: {}",
							bean.getClass().getName(), field.getName(), e.getMessage());
				}
			}
		}
	}
}
