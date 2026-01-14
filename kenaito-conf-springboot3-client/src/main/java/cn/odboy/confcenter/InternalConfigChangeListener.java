package cn.odboy.confcenter;

import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Internal Config Change Listener
 * Polls for config changes via HTTP
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-11
 */
public class InternalConfigChangeListener {

	private final static Logger logger = LoggerFactory.getLogger("cn.odboy.confcenter");

	private static final String LATEST_FETCH_TIMESTAMP = "latest.fetch.timestamp";

	private static final ObjectMapper objectMapper = new ObjectMapper();

	private ScheduledExecutorService hbScheduledExecutor;

	public InternalConfigChangeListener() {
		ConfigcenterContext context = ConfigcenterContext.getInstance();
		registerHttpListener(context);
	}

	public void close() {
		if (hbScheduledExecutor != null) {
			hbScheduledExecutor.shutdown();
		}
	}

	private void registerHttpListener(ConfigcenterContext context) {
		hbScheduledExecutor = Executors.newScheduledThreadPool(1);

		final String url = context.getApiBaseUrls()[0] + "/api/fetchChangedConfigs";
		final Map<String, String> params = new HashMap<>();
		params.put("version", context.getVersion());
		params.put("appName", context.getApp());
		params.put("env", context.getEnv());
		if (context.isIgnoreGlobal()) {
			params.put("ignoreGlobal", String.valueOf(context.isIgnoreGlobal()));
		}
		params.put("lastTime", String.valueOf(System.currentTimeMillis()));

		hbScheduledExecutor.scheduleAtFixedRate(() -> {
			try {
				String safeUrl = context.buildTokenParameter(url);
				String response = HttpUtils.postJson(safeUrl, objectMapper.writeValueAsString(params));

				if (response != null) {
					@SuppressWarnings("unchecked")
					Map<String, Object> map = objectMapper.readValue(response, Map.class);
					if (map.containsKey("data")) {
						@SuppressWarnings("unchecked")
						Map<String, Object> data = (Map<String, Object>) map.get("data");
						if (data.containsKey(LATEST_FETCH_TIMESTAMP)) {
							params.put("lastTime", data.remove(LATEST_FETCH_TIMESTAMP).toString());
						}
						if (!data.isEmpty()) {
							context.updateConfig(data);
						}
					}
				}
			} catch (Exception e) {
				logger.error("Error fetching config changes", e);
			}
		}, context.getSyncIntervalSeconds(), context.getSyncIntervalSeconds(), TimeUnit.SECONDS);

		logger.info("registerHttpUpdaterListener OK intervalSeconds:{}", context.getSyncIntervalSeconds());
	}
}
