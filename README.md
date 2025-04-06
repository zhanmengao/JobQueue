# JobQueue

这是一个go版本工作队列的实现，go中协程的成本已经很低，不再需要对协程进行复用。

DynQueue使用key对请求进行排序，遵守先来后到。对请求大小进行了限制，避免单一用户大量请求导致服务器崩坏。

# cpp
记录了我写过的传统线程池、工作队列实现。本质是生产者消费者模型，供参考。