package cn.odboy.demo;

import java.util.ArrayDeque;
import java.util.ArrayList;
import java.util.Deque;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Component;

import cn.odboy.confcenter.ConfigChangeHanlder;

/**
 * Example {@link ConfigChangeHanlder} that logs changes and keeps the 20 most
 * recent ones for the {@code /changes} endpoint.
 *
 * @author odboy
 * @date 2026-10-21
 */
@Component
public class ConfigChangeLogger implements ConfigChangeHanlder {

    private static final Logger logger = LoggerFactory.getLogger(ConfigChangeLogger.class);

    private static final int MAX_HISTORY = 20;

    private final Deque<Map<String, Object>> history = new ArrayDeque<>();

    @Override
    public synchronized void onConfigChanged(Map<String, Object> changedConfigs) {
        logger.info("[Demo] config changed: {}", changedConfigs);
        history.addFirst(new LinkedHashMap<>(changedConfigs));
        while (history.size() > MAX_HISTORY) {
            history.removeLast();
        }
    }

    public synchronized List<Map<String, Object>> getHistory() {
        return new ArrayList<>(history);
    }
}
