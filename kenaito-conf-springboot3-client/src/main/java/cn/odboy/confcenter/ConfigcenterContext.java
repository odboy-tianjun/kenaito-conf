package cn.odboy.confcenter;

import java.net.InetAddress;
import java.nio.charset.StandardCharsets;
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

import javax.crypto.Cipher;
import javax.crypto.SecretKey;
import javax.crypto.spec.IvParameterSpec;
import javax.crypto.spec.SecretKeySpec;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.core.env.MapPropertySource;
import org.springframework.core.env.MutablePropertySources;
import org.springframework.core.env.PropertiesPropertySource;
import org.springframework.core.env.StandardEnvironment;

import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Config Center Context for Spring Boot 3
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-11
 */
public class ConfigcenterContext {

	private final static Logger logger = LoggerFactory.getLogger("cn.odboy");

	private static ConfigcenterContext instance = new ConfigcenterContext();

	private List<String> sensitiveKeys = new ArrayList<>(Arrays.asList("password","key","secret","token","credentials"));

	public static final String MANAGER_PROPERTY_SOURCE = "kenaitoConfigPropertySource";

	private static final String IGNORE_PLACEHOLER = "[Ignore]";

	private static final String CRYPT_PREFIX = "{Cipher}";

	private static final ObjectMapper objectMapper = new ObjectMapper();

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

	private Properties localProperties = new Properties();

	private StandardEnvironment environment;

	private ConfigcenterContext() {
		String nodeName;
		try {
			nodeName = InetAddress.getLocalHost().getHostName();
		} catch (Exception e) {
			nodeName = "unknown";
		}
		this.nodeId = nodeName;
	}

	public boolean isProcessed() {
		return processed;
	}

	public synchronized void init(Properties properties, boolean isSpringboot) {
		if (processed || !isRemoteEnabled()) return;
		if (properties != null) {
			localProperties.putAll(properties);
		}

		System.setProperty("client.nodeId", nodeId);
		System.setProperty("springboot", String.valueOf(isSpringboot));
		this.isSpringboot = isSpringboot;

		String defaultAppName = getProperty("spring.application.name", "unknown");
		app = getProperty("kenaito.config.appName", defaultAppName);
		if (remoteEnabled == null) {
			remoteEnabled = Boolean.parseBoolean(getProperty("kenaito.config.enabled", "true"));
		}

		if (!isRemoteEnabled()) return;

		env = getProperty("kenaito.config.profile", "dev");

		if (env == null || env.isBlank()) {
			throw new IllegalArgumentException("[kenaito.config.profile] is required");
		}

		setApiBaseUrl(getProperty("kenaito.config.baseUrl", getProperty("kenaito.config.baseUrl", "")));

		version = getProperty("kenaito.config.version", "latest");
		ignoreGlobal = Boolean.parseBoolean(getProperty("kenaito.config.global-ignore", "false"));
		syncIntervalSeconds = Integer.parseInt(getProperty("kenaito.config.sync-interval-seconds", "60"));
		token = getProperty("kenaito.config.token", "");

		System.out.println(String.format("\n=====Configcenter config=====\nappName:%s\nenv:%s\nversion:%s\nremoteEnabled:%s\napiBaseUrls:%s\n=====Configcenter config=====",
				app, env, version, isRemoteEnabled(), toJson(apiBaseUrls)));
	}

	public static ConfigcenterContext getInstance() {
		return instance;
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
		if (apiBaseUrl == null || apiBaseUrl.isBlank()) {
			throw new IllegalArgumentException("[kenaito.config.baseUrl] is required");
		}

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

	public String getSecret() {
		return secret;
	}

	public boolean isRemoteFirst() {
		return remoteFirst;
	}

	public int getSyncIntervalSeconds() {
		return syncIntervalSeconds;
	}

	public String getNodeId() {
		return nodeId;
	}

	public boolean isSpringboot() {
		return isSpringboot;
	}

	public void setEnvironment(StandardEnvironment environment) {
		this.environment = environment;
	}

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

		// Sync merged properties into localProperties for change tracking
		localProperties.putAll(properties);

		// Close existing listener before creating new one to prevent scheduler leak
		if (configChangeListener != null) {
			configChangeListener.close();
		}
		// Register listener
		configChangeListener = new InternalConfigChangeListener();
		printConfigs(properties);
	}

	private Properties getAllRemoteProperties() {
		if (remoteProperties != null) return remoteProperties;
		Properties properties = new Properties();

		Map<String, Object> map = fetchConfigFromServer();
		if (map == null) {
			logger.warn("fetch remote config error and no local cache available, starting with empty remote config");
			return properties;
		}

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
			System.out.println("fetch configs url:" + url);

			try {
				String jsonString = HttpUtils.get(url);
				if (jsonString != null) {
					result = objectMapper.readValue(jsonString, Map.class);
					if (result.containsKey("code")) {
						errorMsg = result.get("msg").toString();
						System.err.println("fetch error:" + errorMsg);
						result = null;
					} else {
						break;
					}
				}
			} catch (Exception e) {
				e.printStackTrace();
			}
		}

		if (result == null) {
			System.out.println(">>>>>remote config fetch error, load from local Cache");
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
			String key = entry.getKey().toString();
			sortKeys.add(key);
		}
		Collections.sort(sortKeys);
		System.out.println("==================final config list start==================");
		String value;
		for (String key : sortKeys) {
			value = hideSensitive(key, properties.getProperty(key));
			System.out.println(String.format("%s = %s", key, value));
		}
		System.out.println("==================final config list end====================");
	}

	public synchronized void updateConfig(Map<String, Object> updateConfig) {
		if (!updateConfig.isEmpty()) {
			Set<String> keySet = updateConfig.keySet();
			for (String key : keySet) {
				String oldValue = localProperties.getProperty(key);
				localProperties.setProperty(key, decodeEncryptIfRequire(updateConfig.get(key)).toString());

				if (environment != null) {
					MutablePropertySources propertySources = environment.getPropertySources();

					synchronized (propertySources) {
						MapPropertySource source = null;
						if (!propertySources.contains(MANAGER_PROPERTY_SOURCE)) {
							source = new MapPropertySource(MANAGER_PROPERTY_SOURCE, new LinkedHashMap<>());
							environment.getPropertySources().addFirst(source);
						} else {
							Object existingSource = propertySources.get(MANAGER_PROPERTY_SOURCE);
							if (existingSource instanceof MapPropertySource) {
								source = (MapPropertySource) existingSource;
							}
						}

						if (source != null) {
							Map<String, Object> map = source.getSource();
							Properties properties = new Properties();
							properties.putAll(map);
							properties.putAll(updateConfig);
							propertySources.replace(source.getName(), new PropertiesPropertySource(source.getName(), properties));
						}
					}
				}

				logger.info("Config [{}] Change, oldValue:{}, newValue:{}", key, oldValue, updateConfig.get(key));
			}

			if (configChangeHanlders == null) {
				configChangeHanlders = new ArrayList<>();
			}

			for (ConfigChangeHanlder handler : configChangeHanlders) {
				try {
					handler.onConfigChanged(updateConfig);
					logger.info("invoke {}.onConfigChanged successed!", handler.getClass().getName());
				} catch (Exception e) {
					e.printStackTrace();
					logger.warn("invoke {}.onConfigChanged error, msg:{}", handler.getClass().getName(), e.getMessage());
				}
			}
		}
	}

	public String buildTokenParameter(String url) {
		if (token == null || token.length() < 8) {
			return url;
		}
		String str = md5Short(token).concat(String.valueOf(new Date().getTime()));
		String authtoken = desEncrypt(token.substring(0, 8), str).toLowerCase();
		return url + (url.contains("?") ? "&" : "?") + "authtoken=" + authtoken;
	}

	public void close() {
		if (configChangeListener != null) {
			configChangeListener.close();
		}
	}

	private Object decodeEncryptIfRequire(Object data) {
		if (data == null) return "";
		if (data.toString().startsWith(CRYPT_PREFIX)) {
			if (secret == null || secret.isBlank()) {
				throw new IllegalArgumentException("config[kenaito.config.encrypt-secret] is required");
			}
			data = data.toString().replace(CRYPT_PREFIX, "");
			String decryptString;
			try {
				decryptString = decryptWithAES(secret, data.toString());
			} catch (Exception e) {
				decryptString = decryptWithAES(globalSecret, data.toString());
			}
			return decryptString;
		}
		return data;
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

	private String hideSensitive(String key, String origin) {
		if (key == null || key.isBlank() || origin == null || origin.isBlank()) return "";
		boolean isSensitive = false;
		for (String k : sensitiveKeys) {
			if (key.toLowerCase().contains(k)) {
				isSensitive = true;
				break;
			}
		}
		int length = origin.length();
		if (isSensitive && length > 1) {
			return origin.substring(0, length / 2).concat("****");
		}
		return origin;
	}

	// Helper methods
	private String getProperty(String key, String defaultValue) {
		String value = localProperties.getProperty(key);
		if (value == null || value.isBlank()) {
			value = System.getProperty(key);
		}
		if (value == null || value.isBlank()) {
			value = System.getenv(key);
		}
		return value != null && !value.isBlank() ? value : defaultValue;
	}

	private String getProperty(String key) {
		return getProperty(key, null);
	}

	private static String toJson(Object obj) {
		try {
			return objectMapper.writeValueAsString(obj);
		} catch (Exception e) {
			return "[]";
		}
	}

	private static String md5(String input) {
		try {
			java.security.MessageDigest md = java.security.MessageDigest.getInstance("MD5");
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

	private static String md5Short(String input) {
		if (input == null || input.isBlank()) return "";
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

	public void addConfigChangeHandler(ConfigChangeHanlder handler) {
		if (configChangeHanlders == null) {
			configChangeHanlders = new ArrayList<>();
		}
		configChangeHanlders.add(handler);
	}
}
