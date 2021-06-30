import * as cdk from "@aws-cdk/core";
import * as ec2 from "@aws-cdk/aws-ec2";
import * as ecs from "@aws-cdk/aws-ecs";

export class NgssamplStack extends cdk.Stack {
  constructor(scope: cdk.Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const vpc = new ec2.Vpc(this, "Vpc", {
      maxAzs: 3, // Default is all AZs in region
    });

    const cluster = new ecs.Cluster(this, "Cluster", {
      vpc: vpc,
    });

    const taskDefinition = new ecs.FargateTaskDefinition(this, "TaskDef");
    taskDefinition.addContainer("Pub", {
      image: ecs.ContainerImage.fromAsset("../"),
      command: ["./app", "-pub", "-creds", "sampler.creds"],
      logging: new ecs.AwsLogDriver({ streamPrefix: "pub" }),
    });
    taskDefinition.addContainer("Sub", {
      image: ecs.ContainerImage.fromAsset("../"),
      command: ["./app", "-sub", "-creds", "sampler.creds"],
      logging: new ecs.AwsLogDriver({ streamPrefix: "sub" }),
    });

    const svc = new ecs.FargateService(this, "Service", {
      cluster,
      taskDefinition,
    });

    svc.connections.addSecurityGroup(
      new ec2.SecurityGroup(this, "OutboundSecGroup", {
        vpc,
        allowAllOutbound: true,
      })
    );
  }
}
