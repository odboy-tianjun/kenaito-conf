package cn.odboy.confcenter;

import java.util.Map;

/**
 * Callback for remote config changes.
 *
 * <p>Implement this interface as a Spring bean to be notified whenever the config
 * center pushes a set of changed keys.</p>
 *
 * @author odboy
 * @version 3.0.0
 * @date 2026-10-21
 */
@FunctionalInterface
public interface ConfigChangeHanlder {

	void onConfigChanged(Map<String, Object> changedConfigs);
}
