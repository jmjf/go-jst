# go-jst

A Golang learning project

## Problem

Foxfire Finance's systems run many batch jobs for business operational and reporting purposes.

A job may include many steps and the output of one job may feed another.

For example, the job that calculates and charges overdrafts at the end of the day produces data that feeds decision support systems. Other jobs use data from the decision support systems to summarize activity so Foxfire can understand overdraft activity and the economic effects it might indicate.

Another job runs every 30 minutes during from 6:00 a.m. to 6:00 p.m. to summarize checking and savings transaction activity for management dashboards. The 6:00 a.m. run summarizes data for each 30 minute period overnight.

To manage expectations, each job has a service level objective (SLO) to deliver data by a certain time.

For example, the overdraft job should produce data to load the decision support system by 3:00 a.m every calendar day. The job that loads the data into the decision support system is due by 5:00 a.m each business day. The report runs on back office business days by 8:00 a.m.

The checking and savings transaction summary and load to dashboard system updates by 10 minutes after and 40 minutes after each hour.

Foxfire wants an application to track job SLO performance to identify:

* Jobs that are late so they can understand consequences and act accordingly
* Jobs that are slowing over time which may cause SLO misses in the future
* Jobs that are routinely early or late to consider changing SLOs or improving infrastructure running the jobs to meet SLOs.
