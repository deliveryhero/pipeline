// Pipeline is a go library that helps you build pipelines without worrying about channel management and concurrency.
// It contains common fan-in and fan-out operations as well as useful utility funcs for batch processing and scaling.
//
// If you have another common use case you would like to see covered by this package, please (open a feature request) https://github.com/deliveryhero/pipeline/issues.
//
// Cookbook
//
// * (How to run a pipeline until the container is killed) https://github.com/deliveryhero/pipeline#PipelineShutsDownWhenContainerIsKilled
// * (How to shut down a pipeline when there is a error) https://github.com/deliveryhero/pipeline#PipelineShutsDownOnError
// * (How to shut down a pipeline after it has finished processing a batch of data) https://github.com/deliveryhero/pipeline#PipelineShutsDownWhenInputChannelIsClosed
//
package pipeline
