Gem::Specification.new do |spec|
  spec.name          = 'kenaito-conf'
  spec.version       = '1.0.3'
  spec.authors       = ['odboy']
  spec.email         = ['odboy@outlook.com']

  spec.summary       = 'Kenaito Config Center Ruby Client'
  spec.description   = 'Ruby client SDK for Kenaito Config Center'
  spec.homepage      = 'https://github.com/odboy/kenaito-conf'
  spec.license       = 'Apache-2.0'
  spec.required_ruby_version = '>= 2.5.0'

  spec.files         = Dir['lib/**/*.rb']
  spec.require_paths = ['lib']

  spec.add_development_dependency 'rake', '~> 13.0'
end
