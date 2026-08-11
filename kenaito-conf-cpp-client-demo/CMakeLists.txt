cmake_minimum_required(VERSION 3.10)
project(kenaito-conf-cpp-client-demo LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# 引入客户端库
add_subdirectory(../kenaito-conf-cpp-client ${CMAKE_BINARY_DIR}/kenaito-conf-cpp-client)

add_executable(demo main.cpp)

target_link_libraries(demo PRIVATE kenaito_conf_client)
