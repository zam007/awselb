pipeline {
    agent {
        node {
            //根据label指定工作的jenkins(在jenkins: Manage Nodes处配置标签)
            label 'closer-jenkins-master'
            //配置工作目录
            customWorkspace "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}/src/${JOB_NAME}"
        }
    }

    //通过tools可自动安装工具，并放置环境变量到PATH
    tools {
        //指定需要的go名称（在jenkins: Global Tool Configuration中配置）
        go 'Go1.10'
    }

    //配置环境变量
    environment {
        GOPATH = "${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_ID}"
        GOBIN = "/usr/local/go/bin"
        GO15VENDOREXPERIMENT = 1
        HTTP_PROXY = "192.168.0.2:8888"
        HTTPS_PROXY = "192.168.0.2:8888"
    }

    stages {
        //从git获取源码
        stage('Checkout SourceCode') {
            steps {
                echo 'Checking code out from git'
                git credentialsId: '3479f8e4-ea9d-44cb-80ae-3b9f1058fbe6', url: 'https://code.tiejin.cn/zam/awselb.git'
            }
        }
        
        //测试前准备，验证dep工具等
        stage('Pre Test Tool') {          
            environment {
                GOPATH = "${JENKINS_HOME}/jobs/"
            }

            steps {
                echo 'Pulling Dependencies'
        
                sh 'go version'
                sh 'go get -u github.com/golang/dep/cmd/dep'
                sh 'go get -u github.com/golang/lint/golint'
                sh 'go get github.com/tebeka/go2xunit'
                

            } 
        }

        //dep vendor install
        stage('Pre Test vendor') {
            steps {
                //or -update
                sh 'dep ensure'                
            }           
        }

        // stage('Test') {
        //     steps {                   
        //         //List all our project files with 'go list ./... | grep -v /vendor/ | grep -v github.com | grep -v golang.org'
        //         //Push our project files relative to ./src
        //         sh 'cd $GOPATH && go list ./... | grep -v /vendor/ | grep -v github.com | grep -v golang.org > projectPaths'
                
        //         //Print them with 'awk '$0="./src/"$0' projectPaths' in order to get full relative path to $GOPATH
        //         def paths = sh returnStdout: true, script: """awk '\$0="./src/"\$0' projectPaths"""
                
        //         echo 'Vetting'

        //         sh """cd $GOPATH && go tool vet ${paths}"""

        //         echo 'Linting'
        //         sh """cd $GOPATH && golint ${paths}"""
                
        //         echo 'Testing'
        //         sh """cd $GOPATH && go test -race -cover ${paths}"""
        //     }
        // }
    
        // stage('Build') {
        //     steps {
        //         echo 'Building Executable'
            
        //         //Produced binary is $GOPATH/src/cmd/project/project
        //         sh """cd $GOPATH/src/cmd/project/ && go build -ldflags '-s'"""
        //     }
        // }
        
        // stage('BitBucket Publish') {
        //     steps {
        //         //Find out commit hash
        //         sh 'git rev-parse HEAD > commit'
        //         def commit = readFile('commit').trim()
            
        //         //Find out current branch
        //         sh 'git name-rev --name-only HEAD > GIT_BRANCH'
        //         def branch = readFile('GIT_BRANCH').trim()
                
        //         //strip off repo-name/origin/ (optional)
        //         branch = branch.substring(branch.lastIndexOf('/') + 1)
            
        //         def archive = "${GOPATH}/project-${branch}-${commit}.tar.gz"

        //         echo "Building Archive ${archive}"
                
        //         sh """tar -cvzf ${archive} $GOPATH/src/cmd/project/project"""

        //         echo "Uploading ${archive} to BitBucket Downloads"
        //         withCredentials([string(credentialsId: 'bb-upload-key', variable: 'KEY')]) { 
        //             sh """curl -s -u 'user:${KEY}' -X POST 'downloads-page-url' --form files=@'${archive}' --fail"""
        //         }
        //     }
        // }
    }
}