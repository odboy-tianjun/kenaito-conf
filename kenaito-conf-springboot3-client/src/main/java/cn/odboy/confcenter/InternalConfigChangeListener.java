package cn.odboy.confcenter;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.alibaba.fastjson2.JSON;

/**
 * Internal config change listener.
 *
 * <p>Polls {@code /api/fetchChangedConfigs} at a fixed interval and applies
 * incremental changes (including deleted keys) to the config center context.</p>
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
public class InternalConfigChangeListener {

	private static final Logger logger = LoggerFactory.getLogger("cn.odboy.confcenter");

	private static final String LATEST_FETCH_TIMESTAMP = "latest.fetch.timestamp";

	private final ScheduledExecutorService scheduler;

	private volatile boolean closed;

	private volatile long lastTime;

	public InternalConfigChangeListener() {
		ConfigcenterContext context = ConfigcenterContext.getInstance();
		this.scheduler = Executors.newSingleThreadScheduledExecutor(runnable -> {
			Thread thread = new Thread(runnable, "kenaito-config-listener");
			thread.setDaemon(true);
			return thread;
		});
		registerHttpListener(context);
	}

	public void close() {
		closed = true;
		scheduler.shutdownNow();
	}

	private void registerHttpListener(ConfigcenterContext context) {
		final String[] baseUrls = context.getApiBaseUrls();
		if (baseUrls.length == 0) {
			logger.warn("[ConfigClient] no config center base url configured, skip change listener");
			return;
		}

		long initialLastTime = context.getLastFetchTimestamp();
		if (initialLastTime <= 0) {
			initialLastTime = System.currentTimeMillis();
		}
		this.lastTime = initialLastTime;
		final int interval = Math.max(1, context.getSyncIntervalSeconds());

		scheduler.scheduleWithFixedDelay(() -> {
			if (closed) {
				return;
			}
			try {
				poll(context, baseUrls);
			} catch (Exception e) {
				logger.error("[ConfigClient] error fetching config changes: {}", e.getMessage(), e);
			}
		}, interval, interval, TimeUnit.SECONDS);

		logger.info("[ConfigClient] change listener registered, intervalSeconds:{}, lastFetchTime:{}",
				interval, initialLastTime);
	}

	@SuppressWarnings("unchecked")
	private void poll(ConfigcenterContext context, String[] baseUrls) {
		for (String baseUrl : baseUrls) {
			String url = context.buildTokenParameter(baseUrl + "/api/fetchChangedConfigs");
			Map<String, Object> params = new HashMap<>();
			params.put("version", context.getVersion());
			params.put("appName", context.getApp());
			params.put("env", context.getEnv());
			params.put("ignoreGlobal", context.isIgnoreGlobal());
			params.put("lastTime", lastTime);

			String response;
			try {
				response = ConfigcenterContext.httpPostJson(url, JSON.toJSONString(params));
			} catch (Exception e) {
				logger.error("[ConfigClient] fetch changed configs error from {}: {}", baseUrl, e.getMessage());
				continue;
			}
			if (response == null) {
				continue;
			}

			Map<String, Object> map = JSON.parseObject(response, Map.class);
			if (map == null) {
				continue;
			}
			Object code = map.get("code");
			// The server returns {"code": 0, "data": {...}} on success and a
			// non-zero code on error. Only non-zero codes are failures.
			if (code instanceof Number codeNumber && codeNumber.intValue() != 0) {
				logger.error("[ConfigClient] fetch changed configs error from {}: {} ({})",
						baseUrl, map.get("msg"), code);
				continue;
			}

			Object dataObject = map.get("data");
			if (!(dataObject instanceof Map)) {
				return;
			}
			Map<String, Object> data = new HashMap<>((Map<String, Object>) dataObject);

			Object timestamp = data.remove(LATEST_FETCH_TIMESTAMP);
			if (timestamp != null) {
				try {
					long parsedTimestamp = Long.parseLong(timestamp.toString());
					context.setLastFetchTimestamp(parsedTimestamp);
					this.lastTime = parsedTimestamp;
				} catch (NumberFormatException e) {
					logger.warn("[ConfigClient] invalid {}: {}", LATEST_FETCH_TIMESTAMP, timestamp);
				}
			}

			Object deleted = data.remove("deletedKeys");
			if (deleted != null) {
				String deletedKeys = deleted.toString();
				if (!deletedKeys.isEmpty()) {
					List<String> keys = new ArrayList<>();
					for (String key : deletedKeys.split(",")) {
						if (!key.isBlank()) {
							keys.add(key.trim());
						}
					}
					if (!keys.isEmpty()) {
						context.removeConfigs(keys);
						logger.info("[ConfigClient] {} configs deleted: {}", keys.size(), keys);
					}
				}
			}

			if (!data.isEmpty()) {
				context.updateConfig(data);
				logger.info("[ConfigClient] {} configs changed: {}", data.size(), data.keySet());
			}
			return;
		}
	}
}
