package cn.odboy.confcenter;

import java.io.IOException;

import org.apache.http.HttpResponse;
import org.apache.http.client.config.RequestConfig;
import org.apache.http.client.methods.HttpGet;
import org.apache.http.client.methods.HttpPost;
import org.apache.http.entity.StringEntity;
import org.apache.http.impl.client.CloseableHttpClient;
import org.apache.http.impl.client.HttpClients;
import org.apache.http.util.EntityUtils;

/**
 * HTTP Client Utilities using Apache HttpClient 4.x
 *
 * @author odboy
 * @version 2.0.0
 * @date 2026-10-11
 */
public class HttpClientUtils {

	private static final CloseableHttpClient httpClient;

	static {
		RequestConfig config = RequestConfig.custom()
				.setConnectTimeout(10000)
				.setSocketTimeout(30000)
				.setConnectionRequestTimeout(5000)
				.build();
		httpClient = HttpClients.custom()
				.setDefaultRequestConfig(config)
				.build();
	}

	public static String get(String url) throws IOException {
		HttpGet request = new HttpGet(url);
		HttpResponse response = httpClient.execute(request);
		int statusCode = response.getStatusLine().getStatusCode();
		String body = EntityUtils.toString(response.getEntity(), "UTF-8");
		if (statusCode >= 200 && statusCode < 300) {
			return body;
		}
		return null;
	}

	public static String postJson(String url, String json) throws IOException {
		HttpPost request = new HttpPost(url);
		request.setHeader("Content-Type", "application/json;charset=UTF-8");
		request.setEntity(new StringEntity(json, "UTF-8"));
		HttpResponse response = httpClient.execute(request);
		int statusCode = response.getStatusLine().getStatusCode();
		String body = EntityUtils.toString(response.getEntity(), "UTF-8");
		if (statusCode >= 200 && statusCode < 300) {
			return body;
		}
		return null;
	}
}
