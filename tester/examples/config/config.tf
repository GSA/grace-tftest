resource "aws_config_configuration_recorder" "config" {
  name     = "config"
  role_arn = "arn:aws:iam::123456789012:role"
  recording_group {
    all_supported                 = true
    include_global_resource_types = true
  }
}

resource "aws_config_delivery_channel" "config" {
  name           = "config"
  s3_bucket_name = "config"
  snapshot_delivery_properties {
    delivery_frequency = "Three_Hours"
  }
  depends_on = [aws_config_configuration_recorder.config]
}

resource "aws_config_configuration_recorder_status" "config" {
  name       = aws_config_configuration_recorder.config.name
  is_enabled = true

  depends_on = [aws_config_delivery_channel.config]
}

/* Not yet supported by moto
resource "aws_config_config_rule" "config" {
  name = "config"
  source {
    owner             = "AWS"
    source_identifier = "CLOUDWATCH_ALARM_ACTION_CHECK"
  }
  depends_on = [aws_config_configuration_recorder.config]
}
*/
