cmake_minimum_required(VERSION 3.10)
project(kenaito-conf-cpp-client VERSION 1.0.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 11)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

find_package(CURL REQUIRED)

# 自动下载 nlohmann/json（单头文件）
set(JSON_INCLUDE_DIR ${CMAKE_CURRENT_SOURCE_DIR}/third_party/nlohmann)
if(NOT EXISTS ${JSON_INCLUDE_DIR}/json.hpp)
    message(STATUS "Downloading nlohmann/json.hpp...")
    file(MAKE_DIRECTORY ${JSON_INCLUDE_DIR})
    file(DOWNLOAD
        https://github.com/nlohmann/json/releases/download/v3.11.3/json.hpp
        ${JSON_INCLUDE_DIR}/json.hpp
        SHOW_PROGRESS
    )
endif()

add_library(kenaito_conf_client STATIC
    src/client.cpp
)

target_include_directories(kenaito_conf_client
    PUBLIC
        ${CMAKE_CURRENT_SOURCE_DIR}/include
        ${CMAKE_CURRENT_SOURCE_DIR}/third_party
    PRIVATE
        ${CURL_INCLUDE_DIRS}
)

target_link_libraries(kenaito_conf_client
    PRIVATE
        ${CURL_LIBRARIES}
)

# Install
install(TARGETS kenaito_conf_client
    ARCHIVE DESTINATION lib
)

install(DIRECTORY include/kenaito_conf
    DESTINATION include
)
