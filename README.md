# Kubelog 

![Logo](https://i.ibb.co/qjLhnxH/c8ae24be-7031-4682-b41d-e314f9f99be7.webp)


## Cheat Sheat
### For pod usage
* Single Pod ```-p```,``` kubelog -p pod_name ```
* Multiple Pod  ```-pm```,``` kubelog -p pod_name1 pod_name1 ```
* Single Pod with search ```-pf```, ```kubelog -pf pod_name "search_text" ```
* Multiple Pod with search ```-fmp``` ,```kubelog -pf pod_name1 pod_name1 "search_text" ```



### For Deployment usage
* Single Deploy ```-d```,``` kubelog -p Deployment_name ```
* Multiple Depoy  ```-mf```,``` kubelog -mf Deployment_name1 Deployment_name2  ```
* Single Deploy with search ```df```, ```kubelog -df Deployment_name1 "search_text" ```
* Multiple Deploy with search ```-fmd``` ,```kubelog -pf Deployment_name1 Deployment_name2 "search_text" ```
