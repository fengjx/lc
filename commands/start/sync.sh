#!/bin/bash

git_url="https://github.com/fengjx/glca.git"
git_branch="start"
remote_template="template_remote"
target_template="template"

# 指定替换的 go module path 变量
module_replace="github\.com\/fengjx\/glca"
module_placeholder="{{.gomod}}"
proj_replace="glca"
proj_placeholder="{{.proj}}"


echo "拉取模板代码"
rm -rf ${remote_template}
git clone -b ${git_branch} ${git_url} ${remote_template}
rm -rf ${remote_template}/.git

# 递归查找并重命名
echo "替换为.tmpl文件"
find "${remote_template}" -type f -name "*.go" -execdir mv {} {}.tmpl \;
find "${remote_template}" -type f -name "*.yml" -execdir mv {} {}.tmpl \;
find "${remote_template}" -type f -name "*.md" -execdir mv {} {}.tmpl \;
find "${remote_template}" -type f -name "*.conf" -execdir mv {} {}.tmpl \;
find "${remote_template}" -type f -name "*.mod" -execdir mv {} {}.tmpl \;
find "${remote_template}" -type f -name "*.service" -execdir mv {} {}.tmpl \;
find "${remote_template}" -type f -name "Makefile" -execdir mv {} {}.tmpl \;
find "${remote_template}" -type f -name "Dockerfile" -execdir mv {} {}.tmpl \;

echo "替换 go module path"
find "${remote_template}" -type f -name "*.tmpl" -execdir sed -i '' "s/$module_replace/${module_placeholder}/g" {} \;

echo "替换 proj name"
find "${remote_template}" -type f -name "*.tmpl" -execdir sed -i '' "s/${proj_replace}/${proj_placeholder}/g" {} \;

rm -rf ${target_template}
mv ${remote_template} ${target_template}

