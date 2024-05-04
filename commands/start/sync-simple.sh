#!/bin/bash

git_url="https://github.com/fengjx/luchen.git"
git_branch="master"
remote_template="template_remote"
remote_template_dir="${remote_template}/_example/quickstart"
target_template="template/simple"

# 指定替换的 go module path 变量
module_replace="github\.com\/fengjx\/luchen\/example\/quickstart"
module_replace2="github\.com\/fengjx\/luchen\/example"
module_placeholder="{{.gomod}}"
proj_replace="quickstart"
proj_placeholder="{{.proj}}"

echo "拉取模板代码"
rm -rf ${remote_template}
git clone -b ${git_branch} ${git_url} ${remote_template}

# 递归查找并重命名
echo "替换为.tmpl文件"
find "${remote_template_dir}" -type f -name "*.go" -execdir mv {} {}.tmpl \;
find "${remote_template_dir}" -type f -name "*.yml" -execdir mv {} {}.tmpl \;
find "${remote_template_dir}" -type f -name "*.md" -execdir mv {} {}.tmpl \;
find "${remote_template_dir}" -type f -name "*.conf" -execdir mv {} {}.tmpl \;
find "${remote_template_dir}" -type f -name "*.service" -execdir mv {} {}.tmpl \;
find "${remote_template_dir}" -type f -name "Makefile" -execdir mv {} {}.tmpl \;
find "${remote_template_dir}" -type f -name "Dockerfile" -execdir mv {} {}.tmpl \;
cp ${remote_template}/_example/go.mod ${remote_template_dir}/go.mod.tmpl

echo "替换 go module path"
find "${remote_template_dir}" -type f -name "*.tmpl" -execdir sed -i '' "s/$module_replace/${module_placeholder}/g" {} \;
find "${remote_template_dir}" -type f -name "*.tmpl" -execdir sed -i '' "s/$module_replace2/${module_placeholder}/g" {} \;

echo "替换 proj name"
find "${remote_template_dir}" -type f -name "*.tmpl" -execdir sed -i '' "s/${proj_replace}/${proj_placeholder}/g" {} \;

rm -rf ${target_template}
mv ${remote_template_dir} ${target_template}
rm -rf ${remote_template}
