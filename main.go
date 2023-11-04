package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
	"github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2transitgateway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		vpcA, err := ec2.NewVpc(ctx, "vpc-a", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.0.0.0/16"),
			EnableDnsHostnames: pulumi.BoolPtr(true),
			EnableDnsSupport:   pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		vpcASubnet, err := ec2.NewSubnet(ctx, "vpc-a-subnet1", &ec2.SubnetArgs{
			CidrBlock: pulumi.String("10.0.0.0/24"),
			VpcId:     vpcA.ID(),
		})

		vpcB, err := ec2.NewVpc(ctx, "vpc-b", &ec2.VpcArgs{
			CidrBlock:          pulumi.String("10.1.0.0/16"),
			EnableDnsHostnames: pulumi.BoolPtr(true),
			EnableDnsSupport:   pulumi.BoolPtr(true),
		})
		if err != nil {
			return err
		}

		vpcBSubnet, err := ec2.NewSubnet(ctx, "vpc-b-subnet1", &ec2.SubnetArgs{
			CidrBlock: pulumi.String("10.1.0.0/24"),
			VpcId:     vpcB.ID(),
		})

		tgw, err := ec2transitgateway.NewTransitGateway(ctx, "tgw", &ec2transitgateway.TransitGatewayArgs{
			DnsSupport:                   pulumi.String("enable"),
			DefaultRouteTableAssociation: pulumi.String("enable"),
			DefaultRouteTablePropagation: pulumi.String("enable"),
			TransitGatewayCidrBlocks:     pulumi.StringArray{pulumi.String("10.2.0.0/24")},
		})
		if err != nil {
			return err
		}

		_, err = ec2transitgateway.NewVpcAttachment(ctx, "attach-vpca", &ec2transitgateway.VpcAttachmentArgs{
			TransitGatewayId: tgw.ID(),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2transitgateway.NewVpcAttachment(ctx, "attach-vpcb", &ec2transitgateway.VpcAttachmentArgs{
			TransitGatewayId: tgw.ID(),
			SubnetIds: pulumi.StringArray{
				vpcBSubnet.ID(),
			},
			VpcId: vpcB.ID(),
		})
		if err != nil {
			return err
		}

		// Create route tables and attach
		rtA, err := ec2.NewRouteTable(ctx, "vpc-a-rt", &ec2.RouteTableArgs{
			VpcId: vpcA.ID(),
			Routes: ec2.RouteTableRouteArray{
				ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: tgw.ID(),
				},
			},
		})
		if err != nil {
			return err
		}
		_, err = ec2.NewRouteTableAssociation(ctx, "vpc-a-rt-attach", &ec2.RouteTableAssociationArgs{
			RouteTableId: rtA.ID(),
			SubnetId:     vpcASubnet.ID(),
		})
		if err != nil {
			return err
		}
		rtb, err := ec2.NewRouteTable(ctx, "vpc-b-rt", &ec2.RouteTableArgs{
			VpcId: vpcB.ID(),
			Routes: ec2.RouteTableRouteArray{
				ec2.RouteTableRouteArgs{
					CidrBlock: pulumi.String("0.0.0.0/0"),
					GatewayId: tgw.ID(),
				},
			},
		})
		if err != nil {
			return err
		}
		_, err = ec2.NewRouteTableAssociation(ctx, "vpc-b-rt-attach", &ec2.RouteTableAssociationArgs{
			RouteTableId: rtb.ID(),
			SubnetId:     vpcBSubnet.ID(),
		})
		if err != nil {
			return err
		}

		vpcASG, err := ec2.NewSecurityGroup(ctx, "vpcA-sg", &ec2.SecurityGroupArgs{
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("traffic from VPCs"),
					FromPort:    pulumi.Int(0),
					ToPort:      pulumi.Int(0),
					Protocol:    pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						vpcA.CidrBlock,
						vpcB.CidrBlock,
					},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						vpcA.CidrBlock,
						vpcB.CidrBlock,
					},
				},
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		vpcBSG, err := ec2.NewSecurityGroup(ctx, "vpcB-sg", &ec2.SecurityGroupArgs{
			Ingress: ec2.SecurityGroupIngressArray{
				&ec2.SecurityGroupIngressArgs{
					Description: pulumi.String("traffic from VPCs"),
					FromPort:    pulumi.Int(0),
					ToPort:      pulumi.Int(0),
					Protocol:    pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						vpcA.CidrBlock,
						vpcB.CidrBlock,
					},
				},
			},
			Egress: ec2.SecurityGroupEgressArray{
				&ec2.SecurityGroupEgressArgs{
					FromPort: pulumi.Int(0),
					ToPort:   pulumi.Int(0),
					Protocol: pulumi.String("-1"),
					CidrBlocks: pulumi.StringArray{
						vpcA.CidrBlock,
						vpcB.CidrBlock,
					},
				},
			},
			VpcId: vpcB.ID(),
		})
		if err != nil {
			return err
		}

		vpcAEndpoint, err := ec2.NewVpcEndpoint(ctx, "vpca-ssm", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ssm"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		vpcBEndpoint, err := ec2.NewVpcEndpoint(ctx, "vpcb-ssm", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ssm"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcBSubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcBSG.ID(),
			},
			VpcId: vpcB.ID(),
		})
		if err != nil {
			return err
		}

		vpcAEndpointEC2Messages, err := ec2.NewVpcEndpoint(ctx, "vpca-ec2messages", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ec2messages"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		vpcBEndpointEC2Messages, err := ec2.NewVpcEndpoint(ctx, "vpcb-ec2messages", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ec2messages"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcBSubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcBSG.ID(),
			},
			VpcId: vpcB.ID(),
		})
		if err != nil {
			return err
		}

		vpcAEndpointSSMMessages, err := ec2.NewVpcEndpoint(ctx, "vpca-ssmmessages", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ssmmessages"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcASubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			VpcId: vpcA.ID(),
		})
		if err != nil {
			return err
		}

		vpcBEndpointSSMMessages, err := ec2.NewVpcEndpoint(ctx, "vpcb-ssmmessages", &ec2.VpcEndpointArgs{
			PrivateDnsEnabled: pulumi.BoolPtr(true),
			ServiceName:       pulumi.String("com.amazonaws.eu-north-1.ssmmessages"),
			VpcEndpointType:   pulumi.String("Interface"),
			SubnetIds: pulumi.StringArray{
				vpcBSubnet.ID(),
			},
			SecurityGroupIds: pulumi.StringArray{
				vpcBSG.ID(),
			},
			VpcId: vpcB.ID(),
		})
		if err != nil {
			return err
		}

		_, err = ec2.NewInstance(ctx, "vpcA-EC2", &ec2.InstanceArgs{
			Ami:      pulumi.String("ami-0b5483e9d9802be1f"),
			SubnetId: vpcASubnet.ID(),
			VpcSecurityGroupIds: pulumi.StringArray{
				vpcASG.ID(),
			},
			InstanceType:       pulumi.String("t4g.nano"),
			IamInstanceProfile: pulumi.String("ec2-ssm-mgmt"),
		}, pulumi.DependsOn([]pulumi.Resource{vpcAEndpoint, vpcAEndpointEC2Messages, vpcAEndpointSSMMessages}))

		_, err = ec2.NewInstance(ctx, "vpcB-EC2", &ec2.InstanceArgs{
			Ami:      pulumi.String("ami-0b5483e9d9802be1f"),
			SubnetId: vpcBSubnet.ID(),
			VpcSecurityGroupIds: pulumi.StringArray{
				vpcBSG.ID(),
			},
			InstanceType:       pulumi.String("t4g.nano"),
			IamInstanceProfile: pulumi.String("ec2-ssm-mgmt"),
		}, pulumi.DependsOn([]pulumi.Resource{vpcBEndpoint, vpcBEndpointEC2Messages, vpcBEndpointSSMMessages}))

		ctx.Export("vpcA", vpcA.ID())
		ctx.Export("vpcB", vpcB.ID())
		return nil
	})
}
