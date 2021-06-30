from influxdb_client import InfluxDBClient, Point
from os import getenv

from influxdb_client.client.write_api import SYNCHRONOUS


def send_data(data_table, lake_prefix, bucket="lakeinfo/autogen", lake_temp=None):
    """Writes data to influxdb client in env properties."""
    client = InfluxDBClient.from_env_properties()
    # client = InfluxDBClient(url=getenv("INFLUXDB_V2_URL"), org=getenv(
    #     "INFLUXDB_V2_ORG"), token=getenv("INFLUXDB_V2_TOKEN"))
    write_api = client.write_api(write_options=SYNCHRONOUS)

    last_point = data_table[-1]
    print(last_point)
    points = [

        Point("{}_level".format(lake_prefix)).tag("units", "ft").field("value", last_point['lake_level']).field(
            "valueNum", float(last_point['lake_level'])),  # .time(last_point['timestamp']),
        Point("{}_turbine_release".format(lake_prefix)).tag("units", "cfps").field(
            "valueNum", last_point['turbine_release_cfs']).field("value", float(last_point['turbine_release_cfs'])),  # .time(last_point['timestamp']),
        Point("{}_spillway_release".format(lake_prefix)).tag("units", "cfps").field(
            "valueNum", last_point['spillway_release_cfs']).field("value", float(last_point['spillway_release_cfs'])),  # .time(last_point['timestamp']),
        Point("{}_total_release".format(lake_prefix)).tag("units", "cfps").field(
            "valueNum", last_point['total_release_cfs']).field("value", float(last_point['total_release_cfs']))  # .time(last_point['timestamp']),
    ]

    if lake_temp:
        points.append(Point("{}_temperature".format(lake_prefix)).tag("units", "ºF").field(
            "valueNum", lake_temp).field("value", lake_temp))

    for i in points:
        write_api.write(bucket, 'patrickjmcd', i)
        print("Wrote {}".format(i._name))

    # query_api = client.query_api()
    # # using Table structure
    # tables = query_api.query(
    #     'from(bucket:"{}") |> range(start: -60m)'.format(bucket), org="patrickjmcd")

    # for table in tables:
    #     print(table)
    #     for row in table.records:
    #         print(row.values)
