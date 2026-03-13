import request from './request'

export function getAppList(params) {
  return request({
    url: '/admin/app/list',
    method: 'post',
    params
  })
}

export function getApp(id) {
  return request({
    url: `/admin/app/${id}`,
    method: 'get'
  })
}

export function addApp(data) {
  return request({
    url: '/admin/app/add',
    method: 'post',
    data
  })
}

export function updateApp(data) {
  return request({
    url: '/admin/app/update',
    method: 'post',
    data
  })
}

export function deleteApp(id) {
  return request({
    url: `/admin/app/delete/${id}`,
    method: 'post'
  })
}

export function getAppOptions() {
  return request({
    url: '/admin/app/options',
    method: 'get'
  })
}

export function generateToken(id) {
  return request({
    url: `/admin/app/generateToken/${id}`,
    method: 'post'
  })
}

export function deleteToken(id) {
  return request({
    url: `/admin/app/deleteToken/${id}`,
    method: 'post'
  })
}
