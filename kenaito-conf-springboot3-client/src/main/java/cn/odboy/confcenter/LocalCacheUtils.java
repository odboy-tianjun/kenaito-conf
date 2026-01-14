package cn.odboy.confcenter;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileInputStream;
import java.io.FileWriter;
import java.io.InputStreamReader;
import java.util.Map;

import com.fasterxml.jackson.databind.ObjectMapper;

/**
 * Local Cache Utilities
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-11
 */
public class LocalCacheUtils {

	private static final ObjectMapper objectMapper = new ObjectMapper();

	private static String localStorageDir = System.getProperty("kenaito.configcenter.local-storage-dir",
			System.getProperty("user.dir"));

	public static void write(Map<String, Object> datas) {
		File dir = new File(localStorageDir);
		if (!dir.exists())
			dir.mkdirs();
		File file = new File(dir, "config-cache.log");
		try (FileWriter fileWriter = new FileWriter(file, false)) {
			fileWriter.write(objectMapper.writeValueAsString(datas));
		} catch (Exception e) {
			e.printStackTrace();
		}
	}

	@SuppressWarnings("unchecked")
	public static Map<String, Object> read() {
		File dir = new File(localStorageDir);
		if (!dir.exists())
			dir.mkdirs();
		File file = new File(dir, "config-cache.log");
		if (!file.exists()) {
			return null;
		}
		try (BufferedReader reader = new BufferedReader(new InputStreamReader(new FileInputStream(file)))) {
			StringBuilder buffer = new StringBuilder();
			String line;
			while ((line = reader.readLine()) != null) {
				buffer.append(line);
			}
			return objectMapper.readValue(buffer.toString(), Map.class);
		} catch (Exception e) {
			e.printStackTrace();
		}
		return null;
	}
}
