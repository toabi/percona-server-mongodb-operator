void CreateCluster(String CLUSTER_PREFIX) {
    withCredentials([string(credentialsId: 'GCP_PROJECT_ID', variable: 'GCP_PROJECT'), file(credentialsId: 'gcloud-key-file', variable: 'CLIENT_SECRET_FILE')]) {
        sh """
            export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
            source $HOME/google-cloud-sdk/path.bash.inc
            gcloud auth activate-service-account --key-file $CLIENT_SECRET_FILE
            gcloud config set project $GCP_PROJECT
            gcloud container clusters create --zone us-central1-a $CLUSTER_NAME-${CLUSTER_PREFIX} --cluster-version 1.15 --machine-type n1-standard-4 --preemptible --num-nodes=3 --network=jenkins-vpc --subnetwork=jenkins-${CLUSTER_PREFIX} --no-enable-autoupgrade
            kubectl create clusterrolebinding cluster-admin-binding --clusterrole cluster-admin --user jenkins@"$GCP_PROJECT".iam.gserviceaccount.com
        """
   }
}
void ShutdownCluster(String CLUSTER_PREFIX) {
    withCredentials([string(credentialsId: 'GCP_PROJECT_ID', variable: 'GCP_PROJECT'), file(credentialsId: 'gcloud-key-file', variable: 'CLIENT_SECRET_FILE')]) {
        sh """
            export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
            source $HOME/google-cloud-sdk/path.bash.inc
            gcloud auth activate-service-account --key-file $CLIENT_SECRET_FILE
            gcloud config set project $GCP_PROJECT
            gcloud container clusters delete --zone us-central1-a $CLUSTER_NAME-${CLUSTER_PREFIX}
        """
   }
}
void pushArtifactFile(String FILE_NAME) {
    echo "Push $FILE_NAME file to S3!"

    withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
        sh """
            touch ${FILE_NAME}
            S3_PATH=s3://percona-jenkins-artifactory/\$JOB_NAME/\$(git rev-parse --short HEAD)
            aws s3 ls \$S3_PATH/${FILE_NAME} || :
            aws s3 cp --quiet ${FILE_NAME} \$S3_PATH/${FILE_NAME} || :
        """
    }
}

void popArtifactFile(String FILE_NAME) {
    echo "Try to get $FILE_NAME file from S3!"

    withCredentials([[$class: 'AmazonWebServicesCredentialsBinding', accessKeyVariable: 'AWS_ACCESS_KEY_ID', credentialsId: 'AMI/OVF', secretKeyVariable: 'AWS_SECRET_ACCESS_KEY']]) {
        sh """
            S3_PATH=s3://percona-jenkins-artifactory/\$JOB_NAME/\$(git rev-parse --short HEAD)
            aws s3 cp --quiet \$S3_PATH/${FILE_NAME} ${FILE_NAME} || :
        """
    }
}

TestsReport = '| Test name  | Status |\\r\\n| ------------- | ------------- |'
testsReportMap  = [:]
testsResultsMap = [:]

void makeReport() {
    for ( test in testsReportMap ) {
        TestsReport = TestsReport + "\\r\\n| ${test.key} | ${test.value} |"
    }
}

void setTestsresults() {
    testsResultsMap.each { file ->
        pushArtifactFile("${file.key}")
    }
}

void runTest(String TEST_NAME, String CLUSTER_PREFIX) {
    def retryCount = 0
    waitUntil {
        try {
            echo "The $TEST_NAME test was started!"
            testsReportMap[TEST_NAME] = 'failed'
            popArtifactFile("${env.GIT_BRANCH}-${env.GIT_SHORT_COMMIT}-$TEST_NAME")

            sh """
                if [ -f "${env.GIT_BRANCH}-${env.GIT_SHORT_COMMIT}-$TEST_NAME" ]; then
                    echo Skip $TEST_NAME test
                else
                    export KUBECONFIG=/tmp/$CLUSTER_NAME-${CLUSTER_PREFIX}
                    source $HOME/google-cloud-sdk/path.bash.inc
                    ./e2e-tests/$TEST_NAME/run
                fi
            """
            testsReportMap[TEST_NAME] = 'passed'
            testsResultsMap["${env.GIT_BRANCH}-${env.GIT_SHORT_COMMIT}-$TEST_NAME"] = 'passed'
            return true
        }
        catch (exc) {
            if (retryCount >= 2) {
                currentBuild.result = 'FAILURE'
                return true
            }
            retryCount++
            return false
        }
    }
    echo "The $TEST_NAME test was finished!"
}

void installRpms() {
    sh '''
        sudo yum install -y https://repo.percona.com/yum/percona-release-latest.noarch.rpm || true
        sudo percona-release enable-only tools
        sudo yum install -y percona-xtrabackup-80 jq | true
    '''
}

def skipBranchBulds = true
if ( env.CHANGE_URL ) {
    skipBranchBulds = false
}

pipeline {
    environment {
        CLOUDSDK_CORE_DISABLE_PROMPTS = 1
        GIT_SHORT_COMMIT = sh(script: 'git describe --always --dirty', , returnStdout: true).trim()
        VERSION = "${env.GIT_BRANCH}-${env.GIT_SHORT_COMMIT}"
        CLUSTER_NAME = sh(script: "echo jenkins-psmdb-${GIT_SHORT_COMMIT} | tr '[:upper:]' '[:lower:]'", , returnStdout: true).trim()
        AUTHOR_NAME  = sh(script: "echo ${CHANGE_AUTHOR_EMAIL} | awk -F'@' '{print \$1}'", , returnStdout: true).trim()
    }
    agent {
        label 'docker'
    }
    stages {
        stage('Prepare') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
                installRpms()
                script {
                    if ( AUTHOR_NAME == 'null' )  {
                        AUTHOR_NAME = sh(script: "git show -s --pretty=%ae | awk -F'@' '{print \$1}'", , returnStdout: true).trim()
                    }
                }
                sh '''
                    if [ ! -d $HOME/google-cloud-sdk/bin ]; then
                        rm -rf $HOME/google-cloud-sdk
                        curl https://sdk.cloud.google.com | bash
                    fi

                    source $HOME/google-cloud-sdk/path.bash.inc
                    gcloud components install alpha
                    gcloud components install kubectl

                    curl -s https://get.helm.sh/helm-v3.2.3-linux-amd64.tar.gz \
                        | sudo tar -C /usr/local/bin --strip-components 1 -zvxpf -
                    curl -s -L https://github.com/openshift/origin/releases/download/v3.11.0/openshift-origin-client-tools-v3.11.0-0cbc58b-linux-64bit.tar.gz \
                        | sudo tar -C /usr/local/bin --strip-components 1 --wildcards -zxvpf - '*/oc'

                    curl -s -L https://github.com/mitchellh/golicense/releases/latest/download/golicense_0.2.0_linux_x86_64.tar.gz \
                        | sudo tar -C /usr/local/bin --wildcards -zxvpf -
                  #  curl -s -L https://github.com/src-d/go-license-detector/releases/latest/download/license-detector.linux_amd64.gz \
                  #      | gunzip | sudo tee /usr/local/bin/license-detector > /dev/null
                    curl -s -L https://github.com/src-d/go-license-detector/releases/download/v3.0.2/license-detector.linux_amd64.gz \
                        | gunzip | sudo tee /usr/local/bin/license-detector > /dev/null
                    sudo chmod +x /usr/local/bin/license-detector

                    sudo sh -c "curl -s -L https://github.com/mikefarah/yq/releases/download/3.3.2/yq_linux_amd64 > /usr/local/bin/yq"
                    sudo chmod +x /usr/local/bin/yq
                '''
                withCredentials([file(credentialsId: 'cloud-secret-file', variable: 'CLOUD_SECRET_FILE')]) {
                    sh '''
                        cp $CLOUD_SECRET_FILE ./e2e-tests/conf/cloud-secret.yml
                    '''
                }
            }
        }
        stage('Build docker image') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
                withCredentials([usernamePassword(credentialsId: 'hub.docker.com', passwordVariable: 'PASS', usernameVariable: 'USER')]) {
                    sh '''
                        DOCKER_TAG=perconalab/percona-server-mongodb-operator:$VERSION
                        docker_tag_file='./results/docker/TAG'
                        mkdir -p $(dirname ${docker_tag_file})
                        echo ${DOCKER_TAG} > "${docker_tag_file}"

                            sg docker -c "
                                docker login -u '${USER}' -p '${PASS}'
                                export RELEASE=0
                                export IMAGE=\$DOCKER_TAG
                                ./e2e-tests/build
                                docker logout
                            "
                            sudo rm -rf ./build
                    '''
                }
                stash includes: 'results/docker/TAG', name: 'IMAGE'
                archiveArtifacts 'results/docker/TAG'
            }
        }
        stage('GoLicenseDetector test') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
               sh """
                   license-detector ${WORKSPACE} | awk '{print \$2}' | awk 'NF > 0' > license-detector-new || true
                   diff -u e2e-tests/license/compare/license-detector license-detector-new
               """
            }
        }
        stage('GoLicense test') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            steps {
                sh '''
                    mkdir -p $WORKSPACE/src/github.com/percona
                    ln -s $WORKSPACE $WORKSPACE/src/github.com/percona/percona-server-mongodb-operator
                    sg docker -c "
                        docker run \
                            --rm \
                            -v $WORKSPACE/src/github.com/percona/percona-server-mongodb-operator:/go/src/github.com/percona/percona-server-mongodb-operator \
                            -w /go/src/github.com/percona/percona-server-mongodb-operator \
                            -e GO111MODULE=on \
                            golang:1.14 sh -c 'go build -v -mod=vendor -o percona-server-mongodb-operator github.com/percona/percona-server-mongodb-operator/cmd/manager'
                    "
                '''

                withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_TOKEN')]) {
                    sh """
                        golicense -plain ./percona-server-mongodb-operator \
                            | grep -v 'license not found' \
                            | sed -r 's/^[^ ]+[ ]+//' \
                            | sort \
                            | uniq \
                            > golicense-new || true

                        diff -u e2e-tests/license/compare/golicense golicense-new
                    """
                }
            }
        }
        stage('Run tests for operator') {
            when {
                expression {
                    !skipBranchBulds
                }
            }
            options {
                timeout(time: 3, unit: 'HOURS')
            }
            parallel {
                stage('E2E Scaling') {
                    steps {
                        CreateCluster('scaling')
                        runTest('init-deploy', 'scaling')
                        runTest('limits', 'scaling')
                        runTest('scaling', 'scaling')
                        runTest('security-context', 'scaling')
                        ShutdownCluster('scaling')
                   }
                }
                stage('E2E Basic Tests') {
                    steps {
                        CreateCluster('basic')
                        runTest('one-pod', 'basic')
                        runTest('monitoring', 'basic')
                        runTest('monitoring-2-0', 'basic')
                        runTest('arbiter', 'basic')
                        runTest('service-per-pod', 'basic')
                        runTest('liveness', 'basic')
                        runTest('users', 'basic')
                        ShutdownCluster('basic')
                    }
                }
                stage('E2E SelfHealing') {
                    steps {
                        CreateCluster('selfhealing')
                        runTest('storage', 'selfhealing')
                        runTest('self-healing', 'selfhealing')
                        runTest('operator-self-healing', 'selfhealing')
                        ShutdownCluster('selfhealing')
                    }
                }
                stage('E2E Backups') {
                    steps {
                        CreateCluster('backups')
                        sleep 60
                        runTest('upgrade', 'backups')
                        runTest('upgrade-consistency', 'backups')
                        runTest('demand-backup', 'backups')
                        runTest('scheduled-backup', 'backups')
                        ShutdownCluster('backups')
                    }
                }
            }
        }
    }
    post {
        always {
            script {
                setTestsresults()
                if (currentBuild.result == null || currentBuild.result == 'SUCCESS') {
                    if (env.CHANGE_URL) {
                        unstash 'IMAGE'
                        def IMAGE = sh(returnStdout: true, script: "cat results/docker/TAG").trim()
                        withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
                            sh """
                                curl -v -X POST \
                                    -H "Authorization: token ${GITHUB_API_TOKEN}" \
                                    -d "{\\"body\\":\\"PSMDB operator docker - ${IMAGE}\\"}" \
                                    "https://api.github.com/repos/\$(echo $CHANGE_URL | cut -d '/' -f 4-5)/issues/${CHANGE_ID}/comments"
                            """
                        }
                    }
                 }
                 else {
                     slackSend channel: '#cloud-dev-ci', color: '#FF0000', message: "[${JOB_NAME}]: build ${currentBuild.result}, ${BUILD_URL} owner: @${AUTHOR_NAME}"
                 }
            }
            script {
                if (env.CHANGE_URL) {
                    withCredentials([string(credentialsId: 'GITHUB_API_TOKEN', variable: 'GITHUB_API_TOKEN')]) {
                        makeReport()
                        sh """
                            curl -v -X POST \
                                -H "Authorization: token ${GITHUB_API_TOKEN}" \
                                -d "{\\"body\\":\\"${TestsReport}\\"}" \
                                "https://api.github.com/repos/\$(echo $CHANGE_URL | cut -d '/' -f 4-5)/issues/${CHANGE_ID}/comments"
                        """
                    }

                    withCredentials([string(credentialsId: 'GCP_PROJECT_ID', variable: 'GCP_PROJECT'), file(credentialsId: 'gcloud-key-file', variable: 'CLIENT_SECRET_FILE')]) {
                        sh '''
                            source $HOME/google-cloud-sdk/path.bash.inc
                            gcloud auth activate-service-account --key-file $CLIENT_SECRET_FILE
                            gcloud config set project $GCP_PROJECT
                            gcloud container clusters delete --zone us-central1-a $CLUSTER_NAME-basic $CLUSTER_NAME-scaling $CLUSTER_NAME-selfhealing $CLUSTER_NAME-backups || true
                            sudo docker rmi -f \$(sudo docker images -q) || true

                            sudo rm -rf $HOME/google-cloud-sdk
                        '''
                    }
                }
            }
            sh '''
                sudo docker rmi -f \$(sudo docker images -q) || true
                sudo rm -rf ./*
                sudo rm -rf $HOME/google-cloud-sdk
            '''
            deleteDir()
        }
    }
}
