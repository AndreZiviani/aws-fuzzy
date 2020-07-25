import click
from pyvis.network import Network

from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import Query
from aws_fuzzy import common


@click.group("plot")
@click.pass_context
def cli(ctx, **kwargs):
    """Plot AWS resources from AWS Config service"""


@cli.command()
@common.p_account()
@common.p_select()
@common.p_region()
@common.p_filter()
@common.p_pager()
@common.p_limit()
@common.p_cache()
@common.p_cache_time()
@common.p_inventory()
@common.p_profile()
@pass_environment
def vpcpeering(ctx, **kwargs):
    """Plot VPC Peering connections graph"""
    kwargs['service'] = "AWS::EC2::VPCPeeringConnection%"
    kwargs['select'] = "configuration.requesterVpcInfo.ownerId" \
                        ", configuration.requesterVpcInfo.vpcId" \
                        ", configuration.accepterVpcInfo.vpcId" \
                        ", configuration.accepterVpcInfo.ownerId" \
                        ", configuration.vpcPeeringConnectionId" \
                        ", tags.tag"
    kwargs['pager'] = False

    query = Query(
        ctx,
        Service=kwargs['service'],
        Select=kwargs['select'],
        Filter=kwargs['filter'],
        Limit=kwargs['limit'],
        Account=kwargs['account'],
        Region=kwargs['region'],
        Pager=kwargs['pager'],
        Cache_time=kwargs['cache_time'],
        Profile=kwargs['profile'])

    if not query.valid:
        query.query()

    ret = query.cached

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
