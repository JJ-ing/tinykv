git基本命令
=====================================
made gy junjie

## 创建仓库

- 在github.com创建（或fork）一个新的仓库获取其url

    例如`https://github.com/JJ-ing/tinykv.git`

- 在本地创建一个空目录，比如 `TinyKv` ，执行命令

        git init    //初始化本地仓库

        git clone https://github.com/JJ-ing/tinykv.git  //克隆到本地

## 日常操作

-  status, add, commit

        git status  //显示当前git仓库状态

        git add abc.txt     //将文件加入暂存区，使其成为git的管理对象

        git add *   //将所有新增或改变的文件加入暂存区

        git commit -m "this is a commit"    //将暂存区的文件实际保存到仓库的离婚斯记录中
        通过这些记录可以在工作树中复原文件

        git commit  //添加详细的commit说明

- push 和 pull 命令

        git push    //推送到远程仓库

        git pull    //从远程仓库拉取到本地

    一般先 pull，再 本地修改，最后 push。


## 更改设置

- 查看和更改远程仓库
    
        git remote -v   //查看远程仓库

        git remote set-url origin <url>     //更改远程仓库地址

        git remote rm origin    //删除远程地址

        git remote add origin [url] //增加远程地址



    