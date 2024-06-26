# DevOps Task

### Demonstration

This project contain controller that satisfy the following requirements:
- Schedule daily scaling changes for specific deployments.
- Schedules are to be saved as Kubernetes CRD objects.
- Controller to:
    1. Translate CRD objects into CronJob objects.
    2. Ensure any updates to CRD objects are reflected on CronJob objects.
        
The controller creates Cronjobs from the CRD objects and ensure the follow:
- create list of cronjobs based on daily schedule (HOUR:MINUTE) and the desired replics
- ensure that any change in any time or the number of replicas will reflect to the cronjobs
- ensure that any deletion in the schedule list will delete the corresponding cronjob
- ensure that control only resources with the Kind ScaleSchedule

The CRD object follow the below example:

```
apiVersion: samplecontroller.k8s.io/v1alpha1
kind: ScaleSchedule
metadata:
  name: example-foo
spec:
  targetRef:
    deploymentName: test
    deploymentNamespace: default
  schedule:
  - at: "22:22"
    replicas: 5
  - at: "07:30"
    replicas: 20
```


The controller is dockerized and pushed to dockerhub

Built Helm chart for deploying CRD, Controller, and Schedule
Objects. 

Created a Jenkinsfile to build, test, and push image.


----------------------------------------------------------------------------

To run the project you can use the Helm Chart in *scaleschedule* folder

```
cd scaleschedule
helm install scaleschedule .
```    


### The following resources will be created
  
  - Controller Deployment
  - CRD
  - CRD objects that will create the Cronjobs

  

### Outputs in Action

![1](https://github.com/Ahmed3sam/scale-schedule/blob/main/screenshots/1.png?raw=true)
![2](https://github.com/Ahmed3sam/scale-schedule/blob/main/screenshots/2.png?raw=true)
![3](https://github.com/Ahmed3sam/scale-schedule/blob/main/screenshots/3.png?raw=true)
![4](https://github.com/Ahmed3sam/scale-schedule/blob/main/screenshots/4.png?raw=true)
![5](https://github.com/Ahmed3sam/scale-schedule/blob/main/screenshots/5.png?raw=true)
![6](https://github.com/Ahmed3sam/scale-schedule/blob/main/screenshots/6.png?raw=true)
![7](https://github.com/Ahmed3sam/scale-schedule/blob/main/screenshots/7.png?raw=true)
