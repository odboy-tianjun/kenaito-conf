package cn.odboy.confcenter;

import java.io.FileReader;
import java.net.InetAddress;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Base64;
import java.util.Collections;
import java.util.Date;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Objects;
import java.util.Properties;
import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;

import javax.crypto.Cipher;
import javax.crypto.SecretKey;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;

import org.apache.commons.lang3.StringUtils;
import org.apache.commons.lang3.Validate;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.ApplicationContext;
import org.springframework.core.env.MapPropertySource;
import org.springframework.core.env.MutablePropertySources;
import org.springframework.core.env.PropertiesPropertySource;
import org.springframework.core.env.StandardEnvironment;

import com.alibaba.fastjson2.JSON;

/**
 * Config Center Context
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-11
 */
public class ConfigcenterContext {

	private final static Logger logger = LoggerFactory.getLogger("cn.odboy");

	private static ConfigcenterContext instance = new ConfigcenterContext();

	private List<String> sensitiveKeys = new ArrayList<>(Arrays.asList("password", "key", "secret", "token", "credentials"));

	public static final String MANAGER_PROPERTY_SOURCE = "kenaitoConfigPropertySource";

	private static final String IGNORE_PLACEHOLER = "[Ignore]";
	private static final String CRYPT_PREFIX = "{Cipher}";

	// Local property storage (replaces ResourceUtils)
	private static final Map<String, String> localProps = new ConcurrentHashMap<>();

	// Spring ApplicationContext (replaces InstanceFactory)
	private static ApplicationContext applicationContext;

	private Boolean remoteEnabled;
	private final String nodeId;
	private String[] apiBaseUrls;
	private String app;
	private String env;
	private String version;
	private boolean ignoreGlobal;
	private String secret;
	private String globalSecret;
	private String token;
	private boolean remoteFirst = false;
	private boolean isSpringboot;
	private int syncIntervalSeconds = 50;
	private InternalConfigChangeListener configChangeListener;
	private List<ConfigChangeHanlder> configChangeHanlders;
	private boolean processed;
	private Properties remoteProperties;

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
		return instance;
	}

	public static void setApplicationContext(ApplicationContext ctx) {
		applicationContext = ctx;
	}

	public static ApplicationContext getApplicationContext() {
		return applicationContext;
	}

	// ==================== Property Utils (replaces ResourceUtils) ====================

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
		if (val == null) val = System.getProperty(key);
		if (val == null) val = System.getenv(key);
		return val;
	}

	public static String getProperty(String key, String defaultValue) {
		String val = getProperty(key);
		return val != null && !val.isEmpty() ? val : defaultValue;
	}

	public static boolean getBoolean(String key, boolean defaultValue) {
		String val = getProperty(key);
		if (val == null || val.isEmpty()) return defaultValue;
		return Boolean.parseBoolean(val);
	}

	public static int getInt(String key, int defaultValue) {
		String val = getProperty(key);
		if (val == null || val.isEmpty()) return defaultValue;
		try {
			return Integer.parseInt(val);
		} catch (NumberFormatException e) {
			return defaultValue;
		}
	}

	public static String getAnyProperty(String... keys) {
		for (String key : keys) {
			String val = getProperty(key);
			if (val != null && !val.isEmpty()) return val;
		}
		return null;
	}

	// ==================== Crypto Utils (replaces mendmix AES/DES/Base64/DigestUtils) ====================

	public static String md5(String input) {
		try {
			MessageDigest md = MessageDigest.getInstance("MD5");
			byte[] digest = md.digest(input.getBytes(StandardCharsets.UTF_8));
			StringBuilder sb = new StringBuilder();
			for (byte b : digest) {
				sb.append(String.format("%02x", b));
			}
			return sb.toString();
		} catch (Exception e) {
			throw new RuntimeException("MD5 error", e);
		}
	}

	public static String md5Short(String input) {
		if (input == null || input.isEmpty()) return "";
		return md5(input).substring(0, 16);
	}

	private static String desEncrypt(String key, String data) {
		try {
			SecretKey secretKey = new SecretKeySpec(key.getBytes(StandardCharsets.UTF_8), "DES");
			Cipher cipher = Cipher.getInstance("DES/ECB/PKCS5Padding");
			cipher.init(Cipher.ENCRYPT_MODE, secretKey);
			byte[] encrypted = cipher.doFinal(data.getBytes(StandardCharsets.UTF_8));
			return Base64.getEncoder().encodeToString(encrypted);
		} catch (Exception e) {
			throw new RuntimeException("DES encrypt error", e);
		}
	}

	private static byte[] aesDecrypt(byte[] data, byte[] key) throws Exception {
		SecretKey secretKey = new SecretKeySpec(key, "AES");
		// Use CBC mode with zero IV to match Go server
		IvParameterSpec iv = new IvParameterSpec(new byte[16]);
		Cipher cipher = Cipher.getInstance("AES/CBC/NoPadding");
		cipher.init(Cipher.DECRYPT_MODE, secretKey, iv);
		byte[] decrypted = cipher.doFinal(data);
		// Remove PKCS7 padding
		int padding = decrypted[decrypted.length - 1] & 0xFF;
		if (padding > 0 && padding <= 16) {
			byte[] result = new byte[decrypted.length - padding];
			System.arraycopy(decrypted, 0, result, 0, result.length);
			return result;
		}
		return decrypted;
	}

	private static String decryptWithAES(String key, String data) {
		try {
			String secretKey = md5(key).substring(16);
			byte[] encrypted = Base64.getDecoder().decode(data.getBytes(StandardCharsets.UTF_8));
			byte[] decrypted = aesDecrypt(encrypted, secretKey.getBytes(StandardCharsets.UTF_8));
			return new String(decrypted, StandardCharsets.UTF_8);
		} catch (Exception e) {
			throw new RuntimeException(String.format("解密错误:%s", data));
		}
	}

	// ==================== HTTP Utils (replaces mendmix HttpUtils) ====================

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
		if (processed || !isRemoteEnabled()) return;
		if (properties != null) mergeProperties(properties);

		System.setProperty("client.nodeId", nodeId);
		System.setProperty("springboot", String.valueOf(isSpringboot));
		this.isSpringboot = isSpringboot;

		String defaultAppName = getProperty("spring.application.name");
		app = getProperty("kenaito.config.appName", defaultAppName);
		if (remoteEnabled == null) remoteEnabled = getBoolean("kenaito.config.enabled", true);

		if (!isRemoteEnabled()) return;

		env = getProperty("kenaito.config.profile", "dev");
		Validate.notBlank(env, "[kenaito.config.profile] is required");

		setApiBaseUrl(getAnyProperty("kenaito.config.baseUrl", "kenaito.config.baseUrl"));

		version = getProperty("kenaito.config.version", "latest");
		ignoreGlobal = getBoolean("kenaito.config.global-ignore", false);
		syncIntervalSeconds = getInt("kenaito.config.sync-interval-seconds", 60);
		token = getProperty("kenaito.config.token");

		logger.info("\n=====Configcenter config=====\nappName:{}\nenv:{}\nversion:{}\nremoteEnabled:{}\napiBaseUrls:{}\n=====Configcenter config=====",
				app, env, version, isRemoteEnabled(), JSON.toJSONString(apiBaseUrls));
	}

	public boolean isRemoteEnabled() {
		return remoteEnabled == null || remoteEnabled;
	}

	public void setRemoteEnabled(boolean remoteEnabled) {
		this.remoteEnabled = remoteEnabled;
	}

	public String[] getApiBaseUrls() {
		return apiBaseUrls;
	}

	public void setApiBaseUrl(String apiBaseUrl) {
		Validate.notBlank(apiBaseUrl, "[kenaito.config.baseUrl] is required");
		String[] urls = apiBaseUrl.split(",|;");
		this.apiBaseUrls = new String[urls.length];
		for (int i = 0; i < urls.length; i++) {
			if (urls[i].endsWith("/")) {
				this.apiBaseUrls[i] = urls[i].substring(0, urls[i].length() - 1);
			} else {
				this.apiBaseUrls[i] = urls[i];
			}
		}
	}

	public String getApp() { return app; }
	public String getEnv() { return env; }
	public String getVersion() { return version; }
	public boolean isIgnoreGlobal() { return ignoreGlobal; }
	public String getSecret() { return secret; }
	public boolean isRemoteFirst() { return remoteFirst; }
	public int getSyncIntervalSeconds() { return syncIntervalSeconds; }
	public String getNodeId() { return nodeId; }
	public boolean isSpringboot() { return isSpringboot; }

	public void mergeRemoteProperties(Properties properties) {
		if (!remoteEnabled) return;
		remoteProperties = getAllRemoteProperties();
		if (remoteProperties != null) {
			Set<Entry<Object, Object>> entrySet = remoteProperties.entrySet();
			for (Entry<Object, Object> entry : entrySet) {
				if (!isRemoteFirst() && properties.containsKey(entry.getKey())) {
					continue;
				}
				String value = entry.getValue().toString();
				if (IGNORE_PLACEHOLER.equals(value)) continue;
				properties.setProperty(entry.getKey().toString(), value);
			}
		}

		// 替换本地变量占位符
		Set<Entry<Object, Object>> entrySet = properties.entrySet();
		for (Entry<Object, Object> entry : entrySet) {
			String key = entry.getKey().toString();
			String value = entry.getValue().toString();
			addProperty(key, value);
		}

		// 加载本地测试配置
		String path = System.getProperty("localtest.config.location");
		if (StringUtils.isNotBlank(path)) {
			try (FileReader fileReader = new FileReader(path)) {
				Properties localProperties = new Properties();
				localProperties.load(fileReader);
				if (!localProperties.isEmpty()) {
					properties.putAll(localProperties);
				}
			} catch (Exception e) {
				logger.warn("Failed to load local test config from {}: {}", path, e.getMessage());
			}
		}

		// Close existing listener before creating new one
		if (configChangeListener != null) {
			configChangeListener.close();
		}
		// register listener
		configChangeListener = new InternalConfigChangeListener();
		printConfigs(properties);
	}

	private Properties getAllRemoteProperties() {
		if (remoteProperties != null) return remoteProperties;
		Properties properties = new Properties();

		Map<String, Object> map = fetchConfigFromServer();
		if (map == null) {
			throw new RuntimeException("fetch remote config error!");
		}

		// 解密密匙
		secret = Objects.toString(map.remove("kenaito.config.encrypt-secret"), null);
		globalSecret = Objects.toString(map.remove("kenaito.config.global-encrypt-secret"), null);
		remoteFirst = Boolean.parseBoolean(Objects.toString(map.remove("kenaito.config.remote-config-first"), "false"));

		Set<String> keys = map.keySet();
		for (String key : keys) {
			Object value = decodeEncryptIfRequire(map.get(key));
			properties.put(key, value);
		}

		return properties;
	}

	@SuppressWarnings("unchecked")
	private Map<String, Object> fetchConfigFromServer() {
		Map<String, Object> result = null;
		String errorMsg = null;
		for (String apiBaseUrl : apiBaseUrls) {
			String url = buildTokenParameter(String.format("%s/api/fetchAllConfigs?appName=%s&env=%s&version=%s&ignoreGlobal=%s",
					apiBaseUrl, app, env, version, ignoreGlobal));
			logger.info("fetch configs url:{}", url);
			try {
				String jsonString = httpGet(url);
				if (jsonString != null) {
					result = JSON.parseObject(jsonString, Map.class);
					if (result.containsKey("code")) {
						errorMsg = result.get("msg").toString();
						logger.error("fetch error:{}", errorMsg);
						result = null;
					} else {
						break;
					}
				}
			} catch (Exception e) {
				logger.error("fetch config error from {}: {}", apiBaseUrl, e.getMessage());
			}
		}

		if (result == null) {
			logger.warn(">>>>>remote config fetch error, load from local Cache");
			result = LocalCacheUtils.read();
		} else {
			LocalCacheUtils.write(result);
		}
		return result;
	}

	private void printConfigs(Properties properties) {
		List<String> sortKeys = new ArrayList<>();
		Set<Entry<Object, Object>> entrySet = properties.entrySet();
		for (Entry<Object, Object> entry : entrySet) {
			sortKeys.add(entry.getKey().toString());
		}
		Collections.sort(sortKeys);
		System.out.println("==================final config list start==================");
		for (String key : sortKeys) {
			String value = hideSensitive(key, properties.getProperty(key));
			System.out.println(String.format("%s = %s", key, value));
		}
		System.out.println("==================final config list end====================");
	}

	@SuppressWarnings("unchecked")
	public synchronized void updateConfig(Map<String, Object> updateConfig) {
		if (!updateConfig.isEmpty()) {
			Set<String> keySet = updateConfig.keySet();
			for (String key : keySet) {
				String oldValue = getProperty(key);
				addProperty(key, decodeEncryptIfRequire(updateConfig.get(key)).toString());
				logger.info("Config [{}] Change, oldValue:{}, newValue:{}", key, oldValue, updateConfig.get(key));
			}

			// Update Spring Environment property source
			if (applicationContext != null) {
				StandardEnvironment environment = (StandardEnvironment) applicationContext.getEnvironment();
				MutablePropertySources propertySources = environment.getPropertySources();
				synchronized (propertySources) {
					Properties properties = new Properties();
					if (propertySources.contains(MANAGER_PROPERTY_SOURCE)) {
						Object source = propertySources.get(MANAGER_PROPERTY_SOURCE);
						if (source instanceof MapPropertySource) {
							Map<String, Object> map = ((MapPropertySource) source).getSource();
							for (Entry<String, Object> entry : map.entrySet()) {
								properties.put(entry.getKey(), entry.getValue());
							}
						} else if (source instanceof PropertiesPropertySource) {
							Map<String, Object> src = ((PropertiesPropertySource) source).getSource();
							for (Entry<String, Object> entry : src.entrySet()) {
								properties.put(entry.getKey(), entry.getValue().toString());
							}
						}
					}
					for (Entry<String, Object> entry : updateConfig.entrySet()) {
						properties.put(entry.getKey(), entry.getValue());
					}
					propertySources.replace(MANAGER_PROPERTY_SOURCE, new PropertiesPropertySource(MANAGER_PROPERTY_SOURCE, properties));
				}
			}

			if (configChangeHanlders == null) {
				configChangeHanlders = new ArrayList<>();
				if (applicationContext != null) {
					Map<String, ConfigChangeHanlder> interfaces = applicationContext.getBeansOfType(ConfigChangeHanlder.class);
					if (interfaces != null) {
						configChangeHanlders.addAll(interfaces.values());
					}
				}
			}

			for (ConfigChangeHanlder handler : configChangeHanlders) {
				try {
					handler.onConfigChanged(updateConfig);
					logger.info("invoke {}.onConfigChanged successed!", handler.getClass().getName());
				} catch (Exception e) {
					logger.warn("invoke {}.onConfigChanged error, msg:{}", handler.getClass().getName(), e.getMessage());
				}
			}
		}
	}

	public boolean pingCcServer(String pingUrl, int retry) {
		boolean result = false;
		try {
			result = httpGet(pingUrl) != null;
		} catch (Exception e) {
			// ignore
		}
		if (retry == 0) return false;
		if (!result) {
			try { Thread.sleep(1500); } catch (Exception e) { }
			return pingCcServer(pingUrl, --retry);
		}
		return result;
	}

	public String buildTokenParameter(String url) {
		if (token == null || token.isEmpty()) {
			return url;
		}
		return url + (url.contains("?") ? "&" : "?") + "token=" + token;
	}

	public void close() {
		if (configChangeListener != null) {
			configChangeListener.close();
		}
	}

	private Object decodeEncryptIfRequire(Object data) {
		if (data == null) return "";
		if (data.toString().startsWith(CRYPT_PREFIX)) {
			Validate.notBlank(secret, "config[kenaito.config.encrypt-secret] is required");
			data = data.toString().replace(CRYPT_PREFIX, "");
			String decryptString;
			try {
				decryptString = decryptWithAES(secret, data.toString());
			} catch (Exception e) {
				if (globalSecret != null && !globalSecret.isEmpty()) {
					decryptString = decryptWithAES(globalSecret, data.toString());
				} else {
					throw e;
				}
			}
			return decryptString;
		}
		return data;
	}

	private String hideSensitive(String key, String origin) {
		if (StringUtils.isAnyBlank(key, origin)) return "";
		boolean isSensitive = false;
		for (String k : sensitiveKeys) {
			if (key.toLowerCase().contains(k)) {
				isSensitive = true;
				break;
			}
		}
		int length = origin.length();
		if (isSensitive && length > 1) return origin.substring(0, length / 2).concat("****");
		return origin;
	}
}
