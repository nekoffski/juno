from conan import ConanFile
from conan.tools.cmake import CMakeToolchain, CMake, cmake_layout, CMakeDeps


class Recipe(ConanFile):
    settings = "os", "compiler", "build_type", "arch"

    def requirements(self):
        self.requires("kstd/1.0")
        self.requires("abseil/20240116.2")
        self.requires("protobuf/5.27.0")
        self.requires("grpc/1.67.1")

        self.tool_requires("protobuf/5.27.0")
        self.tool_requires("grpc/1.67.1")

    def layout(self):
        cmake_layout(self)

    def generate(self):
        deps = CMakeDeps(self)
        deps.generate()
        tc = CMakeToolchain(self)
        tc.generate()

    def build(self):
        cmake = CMake(self)
        cmake.configure()
        cmake.build()
