package cn.odboy.confcenter;

import java.util.Map;

public interface ConfigChangeHanlder {

    void onConfigChanged(Map<String, Object> changedConfigs);
}
