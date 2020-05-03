cmake_minimum_required(VERSION 3.10)
project(verushash)

set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -std=c++11 -D_GLIBCXX_USE_C99 -D_GLIBCXX_USE_C99_MATH") # -Wall
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include)
add_library(verushash STATIC
        src/haraka.c
        src/haraka_portable.c
        src/uint256.cpp
        src/utilstrencodings.cpp
        src/verus_hash.cpp
        src/verus_clhash.cpp
        src/verus_clhash_portable.cpp
        )

########
# LINUX
if(CMAKE_COMPILER_IS_GNUCXX)
    # use native cpu features
    set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -march=x86-64 -mpclmul -msse2 -msse3 -mssse3 -msse4 -msse4.1 -msse4.2 -maes -g -fomit-frame-pointer -fPIC")
    set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -march=x86-64 -mpclmul -msse2 -msse3 -mssse3 -msse4 -msse4.1 -msse4.2 -maes -g -fomit-frame-pointer -fPIC")

    # optimizations
    add_definitions(-O2)

endif()

# MACOS
if(APPLE)
    if(CMAKE_COMPILER_IS_GNUCXX)
        # use native cpu features
        set(CMAKE_C_FLAGS "${CMAKE_C_FLAGS} -m64 -mpclmul -msse2 -msse3 -mssse3 -msse4 -msse4.1 -msse4.2 -maes -g -fomit-frame-pointer -fPIC" )
        set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -m64 -mpclmul -msse2 -msse3 -mssse3 -msse4 -msse4.1 -msse4.2 -maes -g -fomit-frame-pointer -fPIC")
        include_directories(${CMAKE_CURRENT_SOURCE_DIR}/compat)
        # optimizations
        add_definitions(-O2)
    endif()
endif(APPLE)

set_source_files_properties(${CMAKE_CURRENT_SOURCE_DIR}/src/verus_clhash_portable.cpp PROPERTIES COMPILE_FLAGS "-mno-sse4 -mno-sse4.1 -mno-sse4.2 -g")
set_source_files_properties(${CMAKE_CURRENT_SOURCE_DIR}/src/haraka_portable.c PROPERTIES COMPILE_FLAGS "-mno-sse4 -mno-sse4.1 -mno-sse4.2 -g")
set_source_files_properties(${CMAKE_CURRENT_SOURCE_DIR}/src/verus_clhash.cpp PROPERTIES COMPILE_FLAGS "-mavx -g")
set_source_files_properties(${CMAKE_CURRENT_SOURCE_DIR}/src/haraka.c PROPERTIES COMPILE_FLAGS "-mavx -g")

# Common
include_directories(${CMAKE_CURRENT_SOURCE_DIR}/include)

# BOOST
#find_package(Threads REQUIRED COMPONENTS)
# compile boost statically
set(Boost_USE_STATIC_LIBS ON)
set(CMAKE_FIND_LIBRARY_SUFFIXES ".a")
#set(BUILD_SHARED_LIBRARIES OFF)
#set(CMAKE_EXE_LINKER_FLAGS "-static-libgcc -static-libstdc++ -static")
find_package(Boost REQUIRED COMPONENTS system)

if (Boost_FOUND)
    # From the official documentation:
    # Add include directories to the build. [...] If the SYSTEM option is given,
    # the compiler will be told the directories are meant as system include
    # directories on some platforms (signalling this setting might achieve effects
    # such as the compiler skipping warnings [...])."
    include_directories (SYSTEM ${Boost_INCLUDE_DIR})

    # From the offical documentation:
    # "Specify directories in which the linker will look for libraries. [...] Note
    # that this command is rarely necessary. Library locations returned by
    # find_package() and find_library() are absolute paths. Pass these absolute
    # library file paths directly to the target_link_libraries() command. CMake
    # will ensure the linker finds them."
    link_directories (${Boost_LIBRARY_DIRS})
else()
    message("Boost_FOUND NOT FOUND")
endif ()

include_directories(${CMAKE_CURRENT_BINARY_DIR}/../)


set(LIBS ${LIBS} ${Boost_LIBRARIES})

message("-- CXXFLAGS: ${CMAKE_CXX_FLAGS}")
message("-- LIBS: ${LIBS}")

target_link_libraries (verushash ${LIBS})
