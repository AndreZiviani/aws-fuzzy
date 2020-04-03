from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import query
from .common import common_params
from .common import cache_params
from .common import query_params
import click
import json
from pyvis.network import Network


@click.group("plot")
@click.pass_context
def cli(ctx, **kwargs):
    """Plot resources from AWS"""


@cli.command()
@common_params()
@cache_params()
@query_params()
@pass_environment
def vpcpeering(ctx, **kwargs):
    """Plot VPC Peering connections graph"""
    kwargs['service'] = f"AWS::EC2::VPCPeeringConnection%"
    kwargs['select'] = "configuration.requesterVpcInfo.ownerId" \
                        ", configuration.requesterVpcInfo.vpcId" \
                        ", configuration.accepterVpcInfo.vpcId" \
                        ", configuration.accepterVpcInfo.ownerId" \
                        ", configuration.vpcPeeringConnectionId" \
                        ", tags.tag"
    kwargs['pager'] = False
    ret = query(ctx, kwargs)

    net = Network(height="750px", width="100%")
    net.barnes_hut()
    net.show_buttons(filter_=['physics'])

    for peer in ret:
        src_id = peer['configuration']['requesterVpcInfo']['ownerId']
        dst_id = peer['configuration']['accepterVpcInfo']['ownerId']
        src_vpc = peer['configuration']['requesterVpcInfo']['vpcId']
        dst_vpc = peer['configuration']['accepterVpcInfo']['vpcId']
        tags = peer['tags']

        tag = []

        for t in tags:
            tag.append(t['tag'])

        net.add_nodes([str(src_vpc), str(dst_vpc)],
                      title=[str(src_id), str(dst_id)],
                      label=[str(src_vpc), str(dst_vpc)])

        net.add_edge(str(src_vpc), str(dst_vpc), title=",".join(tag))

    net.show("mygraph.html")
    ctx.log("Graph saved to './mygraph.html'")
