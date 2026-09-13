package cn.odboy.confcenter;

import java.io.FileReader;
import java.net.InetAddress;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.HashSet;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Objects;
import java.util.Properties;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;

import org.apache.commons.lang3.StringUtils;
import org.apache.commons.lang3.Validate;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.ApplicationContext;
import org.springframework.core.env.ConfigurableEnvironment;
import org.springframework.core.env.MapPropertySource;
import org.springframework.core.env.MutablePropertySources;
import org.springframework.core.env.PropertiesPropertySource;

import com.alibaba.fastjson2.JSON;

/**
 * Config center context.
 *
 * <p>Holds the client configuration, the remote configuration snapshot and the
 * Spring environment wiring used to refresh {@code @Value} and
 * {@code @ConfigurationProperties} fields at runtime.</p>
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
public class ConfigcenterContext {

	private static final Logger logger = LoggerFactory.getLogger("cn.odboy");

	private static final ConfigcenterContext INSTANCE = new ConfigcenterContext();

	public static final String MANAGER_PROPERTY_SOURCE = "kenaitoConfigPropertySource";

	private static final String IGNORE_PLACEHOLDER = "[Ignore]";

	private static final String LATEST_FETCH_TIMESTAMP = "latest.fetch.timestamp";

	private static final List<String> DEFAULT_SENSITIVE_KEYS = Arrays.asList(
			"password", "passwd", "key", "secret", "token", "credentials");

	/** Properties visible to the client (application config + remote config). */
	private static final Map<String, String> localProps = new ConcurrentHashMap<>();

	private static volatile ApplicationContext applicationContext;

	private static volatile RefreshScopeRegistry refreshScopeRegistry;

	private final String nodeId;

	private Boolean remoteEnabled;

	private String[] apiBaseUrls;

	private String app;

	private String env;

	private String version;

	private boolean ignoreGlobal;

	private String secret;

	private String globalSecret;

	private String token;

	private boolean remoteFirst;

	private boolean isSpringboot;

	private int syncIntervalSeconds = 60;

	private List<String> sensitiveKeys = DEFAULT_SENSITIVE_KEYS;

	private InternalConfigChangeListener configChangeListener;

	private List<ConfigChangeHanlder> configChangeHanlders;

	private boolean processed;

	private volatile Properties remoteProperties;

	private volatile long lastFetchTimestamp;

	private ConfigcenterContext() {
		String name;
		try {
			name = InetAddress.getLocalHost().getHostName();
		} catch (Exception e) {
			name = "unknown";
		}
		this.nodeId = name;
	}

	public static ConfigcenterContext getInstance() {
		return INSTANCE;
	}

	public static void setApplicationContext(ApplicationContext ctx) {
		applicationContext = ctx;
	}

	public static ApplicationContext getApplicationContext() {
		return applicationContext;
	}

	public static void setRefreshScopeRegistry(RefreshScopeRegistry registry) {
		refreshScopeRegistry = registry;
	}

	public static RefreshScopeRegistry getRefreshScopeRegistry() {
		return refreshScopeRegistry;
	}

	// ==================== Property Utilities ====================

	public static void mergeProperties(Properties props) {
		if (props != null) {
			for (Entry<Object, Object> entry : props.entrySet()) {
				localProps.put(entry.getKey().toString(), entry.getValue().toString());
			}
		}
	}

	public static void addProperty(String key, String value) {
		localProps.put(key, value);
	}

	public static String getProperty(String key) {
		String val = localProps.get(key);
		if (val == null) {
			val = System.getProperty(key);
		}
		if (val == null) {
			val = System.getenv(key);
		}
		return val;
	}

	public static String getProperty(String key, String defaultValue) {
		String val = getProperty(key);
		return val != null && !val.isEmpty() ? val : defaultValue;
	}

	public static boolean getBoolean(String key, boolean defaultValue) {
		String val = getProperty(key);
		if (val == null || val.isEmpty()) {
			return defaultValue;
		}
		return Boolean.parseBoolean(val);
	}

	public static int getInt(String key, int defaultValue) {
		String val = getProperty(key);
		if (val == null || val.isEmpty()) {
			return defaultValue;
		}
		try {
			return Integer.parseInt(val.trim());
		} catch (NumberFormatException e) {
			return defaultValue;
		}
	}

	public static String getAnyProperty(String... keys) {
		for (String key : keys) {
			String val = getProperty(key);
			if (val != null && !val.isEmpty()) {
				return val;
			}
		}
		return null;
	}

	static String firstNonBlank(String... values) {
		for (String value : values) {
			if (value != null && !value.isBlank()) {
				return value;
			}
		}
		return null;
	}

	static boolean getBooleanAny(boolean defaultValue, String... keys) {
		String val = getAnyProperty(keys);
		return val == null ? defaultValue : Boolean.parseBoolean(val.trim());
	}

	static int getIntAny(int defaultValue, String... keys) {
		String val = getAnyProperty(keys);
		if (val == null) {
			return defaultValue;
		}
		try {
			return Integer.parseInt(val.trim());
		} catch (NumberFormatException e) {
			return defaultValue;
		}
	}

	// ==================== HTTP Utilities ====================

	public static String httpGet(String url) throws Exception {
		return HttpClientUtils.get(url);
	}

	public static String httpPostJson(String url, String json) throws Exception {
		return HttpClientUtils.postJson(url, json);
	}

	// ==================== Core Methods ====================

	public boolean isProcessed() {
		return processed;
	}

	public synchronized void init(Properties properties, boolean isSpringboot) {
		if (processed) {
			return;
		}
		processed = true;
		if (properties != null) {
			mergeProperties(properties);
		}

		this.isSpringboot = isSpringboot;
		System.setProperty("client.nodeId", nodeId);
		System.setProperty("springboot", String.valueOf(isSpringboot));

		if (remoteEnabled == null) {
			remoteEnabled = getBooleanAny(true,
					"kenaito.config.enabled", "kenaito.configcenter.enabled");
		}
		if (!isRemoteEnabled()) {
			logger.info("[ConfigClient] remote config is disabled");
			return;
		}

		app = firstNonBlank(
				getProperty("kenaito.config.app-name"),
				getProperty("kenaito.config.appName"),
				getProperty("kenaito.configcenter.app-name"),
				getProperty("kenaito.configcenter.appName"),
				getProperty("spring.application.name"));
		Validate.notBlank(app, "[kenaito.config.app-name] is required");

		env = firstNonBlank(
				getProperty("kenaito.config.profile"),
				getProperty("kenaito.config.env"),
				getProperty("kenaito.configcenter.profile"),
				getProperty("kenaito.configcenter.env"),
				"dev");
		Validate.notBlank(env, "[kenaito.config.profile] is required");

		String baseUrl = firstNonBlank(
				getProperty("kenaito.config.base-url"),
				getProperty("kenaito.config.baseUrl"),
				getProperty("kenaito.configcenter.base-url"),
				getProperty("kenaito.configcenter.baseUrl"));
		setApiBaseUrl(baseUrl);

		version = firstNonBlank(
				getProperty("kenaito.config.version"),
				getProperty("kenaito.configcenter.version"),
				"latest");
		ignoreGlobal = getBooleanAny(false,
				"kenaito.config.global-ignore", "kenaito.config.ignore-global",
				"kenaito.configcenter.global-ignore");
		syncIntervalSeconds = getIntAny(60,
				"kenaito.config.sync-interval-seconds", "kenaito.config.sync-interval",
				"kenaito.configcenter.sync-interval-seconds", "kenaito.configcenter.sync-interval");
		token = firstNonBlank(
				getProperty("kenaito.config.token"),
				getProperty("kenaito.configcenter.token"));
		String configuredSensitiveKeys = getAnyProperty(
				"kenaito.config.sensitive-keys", "kenaito.configcenter.sensitive-keys");
		if (StringUtils.isNotBlank(configuredSensitiveKeys)) {
			List<String> keys = new ArrayList<>();
			for (String key : configuredSensitiveKeys.split(",")) {
				if (StringUtils.isNotBlank(key)) {
					keys.add(key.trim().toLowerCase());
				}
			}
			if (!keys.isEmpty()) {
				sensitiveKeys = Collections.unmodifiableList(keys);
			}
		}

		logger.info("\n=====Configcenter config=====\nappName:{}\nenv:{}\nversion:{}\nremoteEnabled:{}\napiBaseUrls:{}\nsyncIntervalSeconds:{}\n=====Configcenter config=====",
				app, env, version, isRemoteEnabled(), JSON.toJSONString(apiBaseUrls), syncIntervalSeconds);
	}

	public boolean isRemoteEnabled() {
		return remoteEnabled == null || remoteEnabled;
	}

	public void setRemoteEnabled(boolean remoteEnabled) {
		this.remoteEnabled = remoteEnabled;
	}

	public String[] getApiBaseUrls() {
		return apiBaseUrls == null ? new String[0] : apiBaseUrls.clone();
	}

	public void setApiBaseUrl(String apiBaseUrl) {
		Validate.notBlank(apiBaseUrl, "[kenaito.config.base-url] is required");
		String[] urls = apiBaseUrl.split("[,;]");
		this.apiBaseUrls = new String[urls.length];
		for (int i = 0; i < urls.length; i++) {
			String url = urls[i].trim();
			if (url.endsWith("/")) {
				url = url.substring(0, url.length() - 1);
			}
			this.apiBaseUrls[i] = url;
		}
	}

	public String getApp() {
		return app;
	}

	public String getEnv() {
		return env;
	}

	public String getVersion() {
		return version;
	}

	public boolean isIgnoreGlobal() {
		return ignoreGlobal;
	}

	public boolean isRemoteFirst() {
		return remoteFirst;
	}

	public boolean isSpringboot() {
		return isSpringboot;
	}

	public int getSyncIntervalSeconds() {
		return syncIntervalSeconds;
	}

	public long getLastFetchTimestamp() {
		return lastFetchTimestamp;
	}

	public void setLastFetchTimestamp(long lastFetchTimestamp) {
		this.lastFetchTimestamp = lastFetchTimestamp;
	}

	public String getNodeId() {
		return nodeId;
	}

	/**
	 * Fetches the remote configuration and injects it into the supplied properties.
	 * Remote fetch failures are logged and never prevent the application from starting.
	 */
	public void mergeRemoteProperties(Properties properties) {
		if (!isRemoteEnabled()) {
			return;
		}
		Properties fetched = getAllRemoteProperties();
		if (fetched != null) {
			for (Entry<Object, Object> entry : fetched.entrySet()) {
				String key = entry.getKey().toString();
				if (!isRemoteFirst() && properties.containsKey(key)) {
					continue;
				}
				String value = Objects.toString(entry.getValue(), "");
				if (IGNORE_PLACEHOLDER.equals(value)) {
					continue;
				}
				properties.setProperty(key, value);
			}
		}

		for (Entry<Object, Object> entry : properties.entrySet()) {
			addProperty(entry.getKey().toString(), entry.getValue().toString());
		}

		// Optional local override file, useful for local debugging.
		String path = System.getProperty("localtest.config.location");
		if (StringUtils.isNotBlank(path)) {
			try (FileReader fileReader = new FileReader(path)) {
				Properties localProperties = new Properties();
				localProperties.load(fileReader);
				if (!localProperties.isEmpty()) {
					properties.putAll(localProperties);
					mergeProperties(localProperties);
				}
			} catch (Exception e) {
				logger.warn("[ConfigClient] failed to load local test config from {}: {}", path, e.getMessage());
			}
		}

		if (configChangeListener != null) {
			configChangeListener.close();
		}
		configChangeListener = new InternalConfigChangeListener();
		printConfigs(properties);
	}

	private Properties getAllRemoteProperties() {
		if (remoteProperties != null) {
			return remoteProperties;
		}
		Map<String, Object> map = fetchConfigFromServer();
		if (map == null) {
			return null;
		}

		secret = Objects.toString(map.remove("kenaito.config.encrypt-secret"), null);
		globalSecret = Objects.toString(map.remove("kenaito.config.global-encrypt-secret"), null);
		remoteFirst = Boolean.parseBoolean(Objects.toString(map.remove("kenaito.config.remote-config-first"), "false"));
		Object timestamp = map.remove(LATEST_FETCH_TIMESTAMP);
		if (timestamp != null) {
			try {
				lastFetchTimestamp = Long.parseLong(timestamp.toString());
			} catch (NumberFormatException e) {
				logger.warn("[ConfigClient] invalid {}: {}", LATEST_FETCH_TIMESTAMP, timestamp);
			}
		}

		Properties properties = new Properties();
		for (Entry<String, Object> entry : map.entrySet()) {
			if (entry.getValue() != null) {
				properties.setProperty(entry.getKey(), entry.getValue().toString());
			}
		}
		remoteProperties = properties;
		return properties;
	}

	@SuppressWarnings("unchecked")
	private Map<String, Object> fetchConfigFromServer() {
		if (apiBaseUrls == null || apiBaseUrls.length == 0) {
			return null;
		}
		Map<String, Object> result = null;
		for (String apiBaseUrl : apiBaseUrls) {
			String url = apiBaseUrl + "/api/fetchAllConfigs";
			Map<String, Object> params = new java.util.HashMap<>();
			params.put("appName", app);
			params.put("env", env);
			params.put("version", version);
			params.put("ignoreGlobal", ignoreGlobal);
			if (StringUtils.isNotBlank(token)) {
				params.put("token", token);
			}
			logger.info("[ConfigClient] fetching configs from {}", url);
			try {
				String jsonString = httpPostJson(url, JSON.toJSONString(params));
				if (jsonString == null) {
					continue;
				}
				Map<String, Object> parsed = JSON.parseObject(jsonString, Map.class);
				// Error responses use {"code": <non-zero number>, "msg": "..."} while
				// successful responses are a plain key/value map (which may legitimately
				// contain a string "code" entry).
				Object code = parsed.get("code");
				if (code instanceof Number codeNumber && codeNumber.intValue() != 0
						&& parsed.get("msg") instanceof String) {
					logger.error("[ConfigClient] fetch config error from {}: {} ({})",
							apiBaseUrl, parsed.get("msg"), code);
					continue;
				}
				result = parsed;
				break;
			} catch (Exception e) {
				logger.error("[ConfigClient] fetch config error from {}: {}", apiBaseUrl, e.getMessage());
			}
		}

		if (result == null) {
			logger.warn("[ConfigClient] remote config fetch failed, loading from local cache");
			result = LocalCacheUtils.read();
		} else {
			LocalCacheUtils.write(result);
			logger.info("[ConfigClient] fetched {} configs", result.size());
		}
		return result;
	}

	private void printConfigs(Properties properties) {
		List<String> sortKeys = new ArrayList<>(properties.stringPropertyNames());
		Collections.sort(sortKeys);
		StringBuilder output = new StringBuilder("\n==================final config list start==================\n");
		for (String key : sortKeys) {
			output.append(key).append(" = ").append(hideSensitive(key, properties.getProperty(key))).append('\n');
		}
		output.append("==================final config list end====================");
		logger.info(output.toString());
	}

	public synchronized void updateConfig(Map<String, Object> updateConfig) {
		if (updateConfig == null || updateConfig.isEmpty()) {
			return;
		}
		for (Entry<String, Object> entry : updateConfig.entrySet()) {
			String key = entry.getKey();
			String newValue = Objects.toString(entry.getValue(), "");
			String oldValue = getProperty(key);
			addProperty(key, newValue);
			if (remoteProperties != null) {
				remoteProperties.setProperty(key, newValue);
			}
			logger.info("[ConfigClient] config [{}] changed, oldValue:{}, newValue:{}",
					key, oldValue, entry.getValue());
		}
		persistRemoteCache();

		updateSpringPropertySource(updateConfig, Collections.emptySet());

		if (configChangeHanlders == null) {
			configChangeHanlders = new ArrayList<>();
			if (applicationContext != null) {
				Map<String, ConfigChangeHanlder> handlers = applicationContext.getBeansOfType(ConfigChangeHanlder.class);
				configChangeHanlders.addAll(handlers.values());
			}
		}
		for (ConfigChangeHanlder handler : configChangeHanlders) {
			try {
				handler.onConfigChanged(updateConfig);
				logger.info("[ConfigClient] invoked {}.onConfigChanged", handler.getClass().getName());
			} catch (Exception e) {
				logger.warn("[ConfigClient] invoke {}.onConfigChanged error: {}", handler.getClass().getName(), e.getMessage());
			}
		}

		refreshDynamicConfigFields(new HashSet<>(updateConfig.keySet()));
	}

	public synchronized void removeConfigs(List<String> keys) {
		if (keys == null || keys.isEmpty()) {
			return;
		}
		for (String key : keys) {
			localProps.remove(key);
			if (remoteProperties != null) {
				remoteProperties.remove(key);
			}
			logger.info("[ConfigClient] config [{}] removed", key);
		}
		persistRemoteCache();
		updateSpringPropertySource(Collections.emptyMap(), new HashSet<>(keys));
		refreshDynamicConfigFields(new HashSet<>(keys));
	}

	private void persistRemoteCache() {
		Properties snapshotSource = remoteProperties;
		if (snapshotSource == null) {
			return;
		}
		Map<String, Object> snapshot = new LinkedHashMap<>();
		for (String name : snapshotSource.stringPropertyNames()) {
			snapshot.put(name, snapshotSource.getProperty(name));
		}
		if (lastFetchTimestamp > 0) {
			snapshot.put(LATEST_FETCH_TIMESTAMP, String.valueOf(lastFetchTimestamp));
		}
		LocalCacheUtils.write(snapshot);
	}

	private void updateSpringPropertySource(Map<String, Object> updates, Set<String> removals) {
		if (applicationContext == null) {
			return;
		}
		if (!(applicationContext.getEnvironment() instanceof ConfigurableEnvironment environment)) {
			return;
		}
		MutablePropertySources propertySources = environment.getPropertySources();
		synchronized (propertySources) {
			Properties properties = new Properties();
			if (propertySources.contains(MANAGER_PROPERTY_SOURCE)) {
				Object source = propertySources.get(MANAGER_PROPERTY_SOURCE);
				if (source instanceof MapPropertySource mapPropertySource) {
					mapPropertySource.getSource().forEach((k, v) -> properties.put(k, Objects.toString(v, "")));
				} else if (source instanceof PropertiesPropertySource propertiesPropertySource) {
					properties.putAll(propertiesPropertySource.getSource());
				}
			}
			for (String key : removals) {
				properties.remove(key);
			}
			updates.forEach((k, v) -> properties.put(k, Objects.toString(v, "")));
			PropertiesPropertySource propertySource =
					new PropertiesPropertySource(MANAGER_PROPERTY_SOURCE, properties);
			if (propertySources.contains(MANAGER_PROPERTY_SOURCE)) {
				propertySources.replace(MANAGER_PROPERTY_SOURCE, propertySource);
			} else {
				propertySources.addFirst(propertySource);
			}
		}
	}

	private void refreshDynamicConfigFields(Set<String> changedKeys) {
		if (applicationContext == null) {
			logger.warn("[RefreshScope] ApplicationContext is null, cannot refresh dynamic config fields");
			return;
		}
		try {
			RefreshScopeRegistry registry = refreshScopeRegistry;
			if (registry == null) {
				// Fallback for environments where the auto-configuration is not active
				// (for example plain Spring applications).
				registry = new RefreshScopeRegistry(applicationContext.getEnvironment());
				for (String beanName : applicationContext.getBeanDefinitionNames()) {
					try {
						registry.postProcessAfterInitialization(applicationContext.getBean(beanName), beanName);
					} catch (Exception ignored) {
						// ignore beans that cannot be resolved
					}
				}
				refreshScopeRegistry = registry;
			}
			int refreshed = registry.refresh(changedKeys);
			if (refreshed > 0) {
				logger.info("[RefreshScope] refreshed {} @Value field(s) for keys: {}", refreshed, changedKeys);
			} else {
				logger.info("[RefreshScope] no registered @Value field references keys {}, "
						+ "their values are already updated in the Spring Environment", changedKeys);
			}
		} catch (Exception e) {
			logger.warn("[RefreshScope] refresh dynamic config fields error: {}", e.getMessage(), e);
		}
	}

	public String getToken() {
		return token;
	}

	public String buildTokenParameter(String url) {
		if (StringUtils.isBlank(token)) {
			return url;
		}
		return url + (url.contains("?") ? "&" : "?") + "token=" + token;
	}

	public void close() {
		if (configChangeListener != null) {
			configChangeListener.close();
		}
	}

	private String hideSensitive(String key, String origin) {
		if (StringUtils.isAnyBlank(key, origin)) {
			return "";
		}
		String lowerKey = key.toLowerCase();
		boolean sensitive = false;
		for (String k : sensitiveKeys) {
			if (lowerKey.contains(k)) {
				sensitive = true;
				break;
			}
		}
		int length = origin.length();
		if (sensitive && length > 1) {
			return origin.substring(0, length / 2) + "****";
		}
		return origin;
	}
}
