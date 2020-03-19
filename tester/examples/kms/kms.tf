data "aws_iam_policy_document" "key" {
  statement {
    effect  = "Allow"
    actions = ["kms:*"]
    principals {
      type        = "AWS"
      identifiers = ["arn:aws:iam::*:root"]
    }
    resources = ["*"]
  }
}

resource "aws_kms_key" "key" {
  deletion_window_in_days = 7
  enable_key_rotation     = true
  policy                  = data.aws_iam_policy_document.key.json
}

resource "aws_kms_alias" "key" {
  name          = "alias/key"
  target_key_id = aws_kms_key.key.key_id
}
