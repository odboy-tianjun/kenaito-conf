# frozen_string_literal: true

Gem::Specification.new do |spec|
  spec.name          = 'kenaito-conf'
  spec.version       = '1.0.0'
  spec.authors       = ['odboy']
  spec.email         = ['tianjun@odboy.cn']

  spec.summary       = 'Kenaito Config Center Ruby Client'
  spec.description   = 'Ruby client for Kenaito Config Center'
  spec.homepage      = 'https://gitea.odboy.cn/odboy/kenaito-conf'
  spec.license       = 'MIT'

  spec.required_ruby_version = '>= 2.7.0'

  spec.files         = Dir['lib/**/*']
  spec.require_paths = ['lib']
end
