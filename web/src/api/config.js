import request from './request'

export function getConfigList(params) {
  return request({
    url: '/admin/config/list',
    method: 'post',
    params
  })
}

export function getConfig(id) {
  return request({
    url: `/admin/config/${id}`,
    method: 'get'
  })
}

export function addConfig(data) {
  return request({
    url: '/admin/config/add',
    method: 'post',
    data
  })
}

export function updateConfig(data) {
  return request({
    url: '/admin/config/update',
    method: 'post',
    data
  })
}

export function deleteConfig(id) {
  return request({
    url: `/admin/config/delete/${id}`,
    method: 'post'
  })
}

export function getConfigHistory(id) {
  return request({
    url: `/admin/config/configHistories/${id}`,
    method: 'get'
  })
}

export function rollbackConfig(id) {
  return request({
    url: `/admin/config/rollback/${id}`,
    method: 'post'
  })
}
