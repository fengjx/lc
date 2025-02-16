#!/bin/bash

set -e

function make_template() {
    if [ $# -lt 1 ]; then
        echo "proj required"
        exit 1
    fi
    proj=$1
    echo "make template ${proj}"
    src_proj="../../template/${proj}"
    target_dir="template/${proj}"

    # 指定替换的 go module path 变量
    module_replace="github\.com\/fengjx\/lc\/${proj}"
    module_placeholder="{{.gomod}}"
    proj_placeholder="{{.proj}}"

    rm -rf ${target_dir}
    cp -r ${src_proj} ${target_dir}

    # 递归查找并重命名
    find "${target_dir}" -type f -name "*.go" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "*.proto" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "*.yml" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "*.md" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "*.conf" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "*.service" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "Makefile" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "Dockerfile" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name "go.mod" -execdir mv {} {}.tmpl \;
    find "${target_dir}" -type f -name ".gitignore" -execdir mv {} {}.tmpl \;

    # 替换 go module path
    find "${target_dir}" -type f -name "*.tmpl" -execdir sed -i '' "s/${module_replace}/${module_placeholder}/g" {} \;

    # 替换 proj name
    find "${target_dir}" -type f -name "*.tmpl" -execdir sed -i '' "s/${proj}/${proj_placeholder}/g" {} \;
}

make_template "simple"
make_template "standard"
