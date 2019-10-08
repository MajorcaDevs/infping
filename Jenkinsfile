@Library('jenkings') _

def getNodeHostname() {
  script {
    sh 'hostname -f'
  }
}

def getGitTag() {
  echo 'trying to get git tag'
  GIT_TAG = sh(script: 'git describe --abbrev=0 --tags', returnStdout: true).trim()
  echo "detected tag : ${GIT_TAG}"
  GIT_COMMIT_DESCRIPTION = sh(script: 'git log -n1 --pretty=format:%B | cat', returnStdout: true).trim()
  echo "current description : ${GIT_COMMIT_DESCRIPTION}"
}

def buildDockerImage(image, arch, tag) {
  script {
    img = docker.build("${image}:${arch}-${tag}", "--pull --build-arg ARCH=${arch} -f docker/Dockerfile .")
  }
}

def pushDockerImage(arch, tag) {
  def githubImage = "docker.pkg.github.com/MajorcaDevs/infping/infping-$arch:$tag"
  script {
    docker.withRegistry('https://registry.hub.docker.com', 'amgxv_dockerhub') { 
      img.push("${arch}-${tag}")  
    }

    docker.withRegistry('https://docker.pkg.github.com', 'amgxv-github-token') {
      sh "docker image tag ${img.id} ${githubImage}"
      docker.image(githubImage).push()  
    }
  }
}

def goBuild(arch, os, gopath) {
  def envVars = ["GOPATH=${gopath}", "GOARCH=${arch}", "GOOS=${os}", "GOCACHE=${gopath}/.go-cache"]
  if(arch == 'arm32v7') {
    envVars.add 'GOARM=7'
  }

  def exeName = "infping-${os}-${arch}"
  withEnv(envVars) {
    sh 'go get -v'
    sh "go build -o ${exeName} -v"
  }

  return exeName
}

def build(arch, os) {
  def gopath = pwd() + '/.go'
  sh "mkdir -p ${gopath}"

  def exeName = goBuild(arch, os, gopath)
  archiveArtifacts artifacts: exeName, fingerprint: true
}

def uploadArtifacts() {
  echo "artifact with tag : ${GIT_TAG}"
  unarchive mapping: ['*': '.']
  githubRelease(
    'amgxv-github-token',
    'amgxv/infping',
    "${GIT_TAG}",
    "infping - Release ${GIT_TAG}",
    "Parse fping output, store result in influxdb - ${GIT_COMMIT_DESCRIPTION}",
    [
      ['infping-linux-amd64', 'application/octet-stream'],
      ['infping-linux-arm', 'application/octet-stream'],
      ['infping-linux-arm64', 'application/octet-stream'],
      ['infping-darwin-amd64', 'application/octet-stream'],
    ]
  )
}


pipeline {
  agent { label '!docker-qemu' }
  environment {
    img = ''
    image = 'amgxv/infping'
    chatId = credentials('amgxv-telegram-chatid')
    GIT_COMMIT_DESCRIPTION = ''
    GIT_TAG = ''
  }

  stages {
    stage('Preparation') {
      steps {
        telegramSend 'infping is being built [here](' + env.BUILD_URL + ')... ', chatId
      }
    }

    stage ('Build Docker Images') {
      parallel {
        stage ('docker-amd64') {
          agent {
            label 'docker-qemu'
          }

          environment {
            arch = 'amd64'
            tag = 'latest'
            img = null
          }

          stages {
            stage ('Current Node') {
              steps {
                getNodeHostname()
              }
            }

            stage ('Build') {
              steps {
                buildDockerImage(image,arch,tag)
              }
            }

            stage ('Push') {
              steps {
                pushDockerImage(arch,tag)
              }
            }
          }
        }

        stage ('docker-arm64') {
          agent {
            label 'docker-qemu'
          }

          environment {
            arch = 'arm64v8'
            tag = 'latest'
            img = null
          }

          stages {
            stage ('Current Node') {
              steps {
                getNodeHostname()
              }
            }

            stage ('Build') {
              steps {
                buildDockerImage(image,arch,tag)
              }
            }

            stage ('Push') {
              steps {
                pushDockerImage(arch,tag)
              }
            }
          }
        }

        stage ('docker-arm32') {
          agent {
            label 'docker-qemu'
          }

          environment {
            arch = 'arm32v7'
            tag = 'latest'
            img = null
          }

          stages {
            stage ('Current Node') {
              steps {
                getNodeHostname()
              }
            }

            stage ('Build') {
              steps {
                buildDockerImage(image,arch,tag)
              }
            }

            stage ('Push') {
              steps {
                pushDockerImage(arch,tag)
              }
            }
          }
        }
      }
    }

    stage('Update manifest (Only Dockerhub)') {
      agent {
        label 'docker-qemu'
      }

      steps {
        script {
          docker.withRegistry('https://index.docker.io/v1/', 'amgxv_dockerhub') {
            getNodeHostname()  
            sh 'docker manifest create amgxv/infping:latest amgxv/infping:amd64-latest amgxv/infping:arm32v7-latest amgxv/infping:arm64v8-latest'
            sh 'docker manifest push -p amgxv/infping:latest'
          }        
        }
      }
    }

    stage ('Build Go Binaries') {
      parallel {
        stage ('amd64') {
          agent {
            docker {
              image 'golang:buster'
            }
          }

          environment {
            os = 'linux'
            arch = 'amd64'
          }

          steps {
            build(arch, os)
          }
        }

        stage ('arm') {
          agent {
            docker {
              image 'golang:buster'
            }
          }

          environment {
            os = 'linux'
            arch = 'arm'
          }

          steps {
            build(arch, os)
          }
        }

        stage ('arm64') {
          agent {
            docker {
              image 'golang:buster'
            }
          }

          environment {
            os = 'linux'
            arch = 'arm64'
          }

          steps {
            build(arch, os)
          }
        }

        stage ('darwin-amd64') {
          agent {
            docker {
              image 'golang:buster'
            }
          }

          environment {
            os = 'darwin'
            arch = 'amd64'
          }

          steps {
            build(arch, os)
          }
        }

        stage ('windows-amd64') {
          agent {
            docker {
              image 'golang:buster'
            }
          }

          environment {
            os = 'windows'
            arch = 'amd64'
          }

          steps {
            build(arch, os)
          }
        }
      }
    }

  stage ('Get Git Tag') {
    agent { label 'majorcadevs' }
    steps {
      getGitTag()
    }
  }

  stage('Upload Release to Github') {
      agent { label 'majorcadevs' }
      when { expression { GIT_TAG != null } }
      steps {
        uploadArtifacts()
      }
    }
  }    
    
  post {
    success {
      cleanWs()
      telegramSend 'infping has been built successfully. ', chatId
    }

    failure {
      cleanWs()
      telegramSend 'infping could not have been built.\n\nSee [build log](' + env.BUILD_URL + ')... ', chatId
    }
  }
}
