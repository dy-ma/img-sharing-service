from diagrams import Diagram, Cluster
from diagrams.aws.compute import Lambda 
from diagrams.aws.network import APIGateway 
from diagrams.aws.network import CloudFront
from diagrams.aws.network import Route53
from diagrams.aws.database import Dynamodb 
from diagrams.aws.storage import S3Glacier
from diagrams.aws.storage import SimpleStorageServiceS3
from diagrams.aws.network import VPC
from diagrams.onprem.client import Client

with Diagram("Image Sharing Service"):
    client = Client("Client")
    dns = Route53("DNS")
    cdn = CloudFront("CDN")
    gateway = APIGateway("API Gateway")

    with Cluster("VPC"):
        handlers = Lambda("Route Handlers (Go)")
        ddb = Dynamodb("Album Metadata")
        s3 = SimpleStorageServiceS3("Images")
        s3g = S3Glacier("Expired Images")

    client >> dns 
    client >> cdn
    cdn - gateway - handlers
    handlers - ddb
    handlers << s3 >> s3g
    handlers << s3g
