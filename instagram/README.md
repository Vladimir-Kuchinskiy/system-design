# Design Instagram

## Step 1: Outline use cases and constraints

### Use cases

#### We'll scope the problem to handle only the following use cases

* **User** views a post 
* **User** views a post and leave a **Like**
* **User** views a post and leave a **Comment**
* **User** views the news feed and can see the timeline of posts of the users he/she follows

#### Out of scope

* **User** registers for an account
    * **User** verifies email
* **User** logs in/out into a registered account
    * **User** edits a profile
* **User** can follow/unfollow other **User**
* **User** can upload a **Post**

### Constraints and assumptions

#### State assumptions

* Traffic is not evenly distributed
* Showing **Timeline** should be fast
* **Timeline** should be always available for the user, availability, and partition tolerance are important.
* Eventual consistency for a **Timeline** is okay.
* **Timeline** update should NOT be in real-time, it is okay to have a 1-5 sec delay.
* 10 million users
* 100 million writes per month (10 mil posts + 30 mil comments + 60 mil likes).
* 1000 million (1 milliard) reads per month.
* 10:1 read-to-write ratio

#### Calculate usage

* Size per Post ~ 10 Mb content per post
* Size per Like ~ uint64 + uint64 + timestamp (64+64+12) = 140 bytes ~ 0.14 Kb
* Size per Comment ~ uint64 + uint64 + timestamp + varchar(1024) = 1164 bytes ~ 1.16 Kb
* 10 Mb Post * 10 million posts         = 100 Tb posts per month
* 0.14 Kb Like * 60 million likes       = 8.4 Gb likes per month
* 1.16 Kb Comment * 30 million comments = 34.8 Gb comments per month
* Assume most are new posts instead of updates to existing ones
* 4 post writes per second on average
* 24 like writes per second on average
* 12 comment writes per second on average
* 400 read requests per second on average

## Step 2: Create a high level design

> Outline a high level design with all important components.

![Imgur](Basic-Insta-Design.png)

## Step 3: Design core components

### Use case: User views timeline

The `Posts` table could have the following structure:

```
id uint64 NOT NULL serial
user_id uint64 NOT NULL
description varchar(1024) DEFAULT NULL
created_at datetime NOT NULL
updated_at datetime
PRIMARY KEY(id)
```
To be able to query user posts by user we need a B-tree index on (user_id) column

The `Comments` table could have the following structure:

```
id uint64 NOT NULL serial
post_id uint64 NOT NULL
user_id uint64 NOT NULL
comment varchar(1024) NOT NULL
created_at datetime NOT NULL
updated_at datetime
PRIMARY KEY(id)
```
To be able to query comments by post we need a B-tree index on (post_id) column
To be able to sort posts by created_at we need a B-tree index on (created_at) column

The `Likes` table could have the following structure:

```
user_id uint64 NOT NULL
post_id uint64 NOT NULL
created_at datetime NOT NULL
PRIMARY KEY(user_id,post_id)
```
To be able to query likes by post we need a B-tree index on (post_id) column

The `Follows` table could have the following structure:

```
followee_id uint64 NOT NULL
follower_id uint64 NOT NULL
created_at datetime NOT NULL
PRIMARY KEY(follower_id,follower_id)
```

To be able to query followees of a user we need a B-tree index on (follower_id) column
To be able to query followers of a user we need a B-tree index on (followee_id) column

To be able to generate a timeline we need to do next steps:
Current user id = 111;
1) Followees => `SELECT * FROM follows WHERE follower_id = 111`
2) Followees user profiles => `SELECT * FROM user_profile WHERE user_id in (followees_ids)`
3) Followees resent Posts => `SELECT * FROM posts WHERE user_id in (followees_ids) ORDER BY created_at DESC LIMIT 20`
4) Posts likes => `SELECT post_id, count(*) FROM likes WHERE post_id in (followees_posts_ids)`
6) Aggregate result

## Step 4: Scale the design

> Identify and address bottlenecks, given the constraints.

![Imgur](Final-Design.png)

### SPOFs

Basic design had lots of SPOFs, all the services and all the databases had a single point of failure.
The solutions were:
1) Introduce a set of **Load Balancers** at the front of a services infrastructure.
2) Scale all **API** services to minimum 2 replicas
3) Scale **Databases** to Multi-AZ setup with a single side replica for recovery.

### Bottlenecks

1) **Blob Store** was a bottleneck in terms of performance, it was affecting the quickness of the Timeline rendering.
The solution is to introduce the CDN that will be caching static content like WebPages, Photos and Videos.
2) **Timeline Service** was a bottleneck because it was querying all the services to Prepare a timeline on each request which were increasing the averall system load and since we will have ~ 400 read requests per sec on average it is not the setup we need.
The solution was to start using cache, and there were couple of caching strategies that we could select:
* **Cache Aside** - in this case we would prepare timeline for each new request by goting to all the services and aggregating the info, and store it in the cache, so the next time request comes we are able to return responce from the cache.
* **Refresh-ahead** - in this case we would precompute the cache reactively on a data change on side, so the **Timeline Service** just need to go to read from the cache.

The second caching option was selected because of 2 reasons:
1) High Availability requirements, so the **Timeline Service** should always return some data, event if it is outdated.
2) Evential Consistency is a good enought option, so again updating timeline in a near real time speed is a good enough solution.

Because **Timeline Service** always will be reading in memory stored data, the read should fit our requirements of ~400 reads per second.
