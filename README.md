#部署方式
1. 把文件夹"front-end"下的内容发送至装有centos7 64位系统的服务器
2. 服务器上执行命令。
`
curl -sL https://rpm.nodesource.com/setup_12.x | bash -
`
命令以安装npm和nodejs的最新版本。
3. 在fileManager文件夹的根目录下运行命令
`
npm run build
`
。
3. 服务器上安装MySQL5.7。
4. 上传backend文件夹下的table_creation.sql至该用户根目录。
5. 建立数据库用户后，登陆该用户至DBMS，在该用户的管理控制台下执行source table_create.sql 。
6. 服务器上安装golang语言编译环境，版本1，12，并配置好各环境变量。
7. 通过如下命令安装代码所需的依赖包：
`
go get github.com/go-sql-driver/mysql ; 
go get github.com/gorilla/mux ; 
go get github.com/gorilla/sessions
` 
8. 修改backend文件夹下的main.go文件中的dbUser和dbPwd变量，使其适应用户环境下的数据库配置。(数据库名不用修改)
9. 上传main.go文件至服务器下的dist文件夹。
10. dist目录下依次运行
`
go build main.go ; ./main
`
以启动网站服务。
11. 通过服务器绑定的域名即可直接访问该网站。