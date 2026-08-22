package cn.odboy.confcenter;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileWriter;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.alibaba.fastjson2.JSON;

/**
 * Local disaster-recovery cache.
 *
 * <p>The last successfully fetched remote configuration is written to
 * {@code config-cache.log} so that the application can still start when the
 * config center is temporarily unavailable.</p>
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
public class LocalCacheUtils {

	private static final Logger logger = LoggerFactory.getLogger(LocalCacheUtils.class);

	private static final String CACHE_FILE_NAME = "config-cache.log";

	private LocalCacheUtils() {
	}

	private static String getStorageDir() {
		String dir = ConfigcenterContext.getAnyProperty(
				"kenaito.config.local-storage-dir",
				"kenaito.configcenter.local-storage-dir");
		return dir != null ? dir : System.getProperty("user.dir");
	}

	public static void write(Map<String, Object> datas) {
		File dir = new File(getStorageDir());
		if (!dir.exists() && !dir.mkdirs()) {
			logger.warn("[ConfigClient] cannot create local cache directory: {}", dir);
			return;
		}
		File file = new File(dir, CACHE_FILE_NAME);
		try (FileWriter fileWriter = new FileWriter(file, StandardCharsets.UTF_8, false)) {
			fileWriter.write(JSON.toJSONString(datas));
		} catch (Exception e) {
			logger.warn("[ConfigClient] write local cache error: {}", e.getMessage());
		}
	}

	@SuppressWarnings("unchecked")
	public static Map<String, Object> read() {
		File dir = new File(getStorageDir());
		if (!dir.exists() && !dir.mkdirs()) {
			return null;
		}
		File file = new File(dir, CACHE_FILE_NAME);
		if (!file.exists()) {
			return null;
		}
		try (BufferedReader reader = new BufferedReader(
				new InputStreamReader(new FileInputStream(file), StandardCharsets.UTF_8))) {
			StringBuilder buffer = new StringBuilder();
			String line;
			while ((line = reader.readLine()) != null) {
				buffer.append(line);
			}
			return JSON.parseObject(buffer.toString(), Map.class);
		} catch (Exception e) {
			logger.warn("[ConfigClient] read local cache error: {}", e.getMessage());
		}
		return null;
	}
}
