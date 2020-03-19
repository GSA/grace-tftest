resource "aws_cloudwatch_event_rule" "rule" {
  name = "rule"

  event_pattern = <<PATTERN
{
  "detail-type": [
    "AWS Console Sign In via CloudTrail"
  ]
}
PATTERN
}

resource "aws_cloudwatch_event_target" "target" {
  rule = aws_cloudwatch_event_rule.rule.name
  arn  = "arn:aws:iam::123456789012:target"
}
