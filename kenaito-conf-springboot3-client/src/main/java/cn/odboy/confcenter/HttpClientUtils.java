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
 * HTTP client utilities based on Apache HttpClient 4.x.
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
public class HttpClientUtils {

	private static final CloseableHttpClient HTTP_CLIENT;

	static {
		RequestConfig config = RequestConfig.custom()
				.setConnectTimeout(10_000)
				.setSocketTimeout(30_000)
				.setConnectionRequestTimeout(5_000)
				.build();
		HTTP_CLIENT = HttpClients.custom()
				.setDefaultRequestConfig(config)
				.build();
	}

	private HttpClientUtils() {
	}

	public static String get(String url) throws IOException {
		HttpGet request = new HttpGet(url);
		return execute(request);
	}

	public static String postJson(String url, String json) throws IOException {
		HttpPost request = new HttpPost(url);
		request.setHeader("Content-Type", "application/json;charset=UTF-8");
		request.setEntity(new StringEntity(json, "UTF-8"));
		return execute(request);
	}

	private static String execute(org.apache.http.client.methods.HttpRequestBase request) throws IOException {
		HttpResponse response = HTTP_CLIENT.execute(request);
		int statusCode = response.getStatusLine().getStatusCode();
		String body = response.getEntity() == null ? null : EntityUtils.toString(response.getEntity(), "UTF-8");
		if (statusCode >= 200 && statusCode < 300) {
			return body;
		}
		return null;
	}
}
