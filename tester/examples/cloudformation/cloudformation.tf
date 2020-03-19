resource "aws_cloudformation_stack" "stack" {
  name = "stack"

  template_body = <<EOF
  {
    "AWSTemplateFormatVersion": "2010-09-09",
    "Resources": {
      "EmailSNSTopic": {
        "Type": "AWS::SNS::Topic",
        "Properties": {
          "DisplayName": "Email Topic",
          "Subscription": [
            {
              "Endpoint": "email@address.com",
              "Protocol": "email"
            }
          ]
        }
      },
    "Outputs": {
      "Arn": {
        "Description": "Email SNS Topic ARN",
        "Value": {
          "Ref": "EmailSNSTopic"
        }
      }
    }
  }
}
EOF
}
