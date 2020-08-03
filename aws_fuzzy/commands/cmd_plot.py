import click
from pyvis.network import Network

from aws_fuzzy.cli import pass_environment
from aws_fuzzy.query import Query
from aws_fuzzy import common


class Plot(common.Cache):
    def __init__(self, ctx, Cache, Cache_time, Directed=False):
        super().__init__(ctx, "plot", Cache_time)

        self.graph = Network(height="100%", width="100%", directed=Directed)
        self.graph.barnes_hut()
        self.graph.show_buttons(filter_=['physics'])


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

    # get information on the peerings
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

    plot = Plot(ctx, False, kwargs['cache_time'])

    for peer in ret:
        src_id = peer['configuration']['requesterVpcInfo']['ownerId']
        dst_id = peer['configuration']['accepterVpcInfo']['ownerId']
        src_vpc = peer['configuration']['requesterVpcInfo']['vpcId']
        dst_vpc = peer['configuration']['accepterVpcInfo']['vpcId']
        tags = peer['tags']

        tag = []

        for t in tags:
            tag.append(t['tag'])

        # get the profile name associated with the account id of the peering connection
        src_name = query.account_ids.get(str(src_id),
                                         {'name': str(src_id)})['name']
        dst_name = query.account_ids.get(str(dst_id),
                                         {'name': str(dst_id)})['name']

        # add nodes to graph
        plot.graph.add_node(
            str(src_vpc),
            title=str(src_name),
            group=str(src_name),
            size=40,
            label=f"{src_name}\n{src_vpc}")
        plot.graph.add_node(
            str(dst_vpc),
            title=str(dst_name),
            group=str(dst_name),
            size=40,
            label=f"{dst_name}\n{dst_vpc}")

        # add edge between nodes
        plot.graph.add_edge(str(src_vpc), str(dst_vpc), title=",".join(tag))

    plot.graph.show("mygraph.html")
    ctx.log("Graph saved to './mygraph.html'")


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
@click.argument('securitygroup')
@pass_environment
def securitygroups(ctx, **kwargs):
    """Plot which SecurityGroups have a relationship with the desired SecurityGroup"""

    plot = Plot(ctx, False, kwargs['cache_time'], Directed=True)

    # get information on the desired sg
    query_sg = Query(
        ctx,
        Service="AWS::EC2::SecurityGroup%",
        Select="resourceId, resourceName, configuration",
        Filter=f"resourceId = '{kwargs['securitygroup']}'",
        Limit=kwargs['limit'],
        Account=kwargs['account'],
        Region=kwargs['region'],
        Pager=False,
        Cache_time=kwargs['cache_time'],
        Profile=kwargs['profile'])

    if not query_sg.valid:
        query_sg.query()

    me = query_sg.cached[0]['configuration']

    # get the profile name associated with the account id of the sg
    account = query_sg.account_ids.get(me['ownerId'],
                                       {'name': me['ownerId']})['name']

    # add node to graph
    plot.graph.add_node(
        me['groupId'],
        title=account,
        group=account,
        size=40,
        label=f"{me['groupName']}\n{me['groupId']}")

    # get all other sg that references the desired sg in egress policy
    query = Query(
        ctx,
        Service="AWS::EC2::SecurityGroup%",
        Select="resourceId, resourceName, configuration",
        Filter=
        f"configuration.ipPermissionsEgress.userIdGroupPairs.groupId = '{kwargs['securitygroup']}'",
        Limit=kwargs['limit'],
        Account=kwargs['account'],
        Region=kwargs['region'],
        Pager=False,
        Cache_time=kwargs['cache_time'],
        Profile=kwargs['profile'])

    if not query.valid:
        query.query()

    ret = query.cached

    #query.print()
    for sg in ret:
        sg_name = sg['resourceName']
        sg_id = sg['resourceId']

        # for each port in this SG
        for rule in sg['configuration']['ipPermissionsEgress']:
            # for each SG associated with this port
            for dest in rule['userIdGroupPairs']:
                # if this rule have the SG we are looking for
                if dest['groupId'] == kwargs['securitygroup']:
                    # get the profile name associated with the account id of the sg
                    account = query.account_ids.get(
                        sg['configuration']['ownerId'],
                        {'name': sg['configuration']['ownerId']})['name']

                    # add node to graph
                    plot.graph.add_node(
                        str(sg_id),
                        title=account,
                        group=account,
                        size=40,
                        label=f"{sg_name}\n{sg_id}")

                    # add edge between nodes
                    plot.graph.add_edge(
                        sg_id,
                        me['groupId'],
                        title=
                        f"{rule['ipProtocol']}: {int(rule['fromPort'])}-{int(rule['toPort'])}"
                    )

    plot.graph.show("mygraph.html")
    ctx.log("Graph saved to './mygraph.html'")
