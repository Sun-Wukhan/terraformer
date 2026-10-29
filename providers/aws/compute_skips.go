package aws

// Compute sweep — types deliberately not implemented, with reasons
// (see results ledger for the full per-type disposition):
//
//   - aws_elastic_beanstalk_application_version,
//     aws_elastic_beanstalk_configuration_template: provider docs have
//     no Import section (no provider importer).
//   - aws_codecommit_trigger: provider docs have no Import section
//     (no provider importer).
//   - aws_codecommit_approval_rule_template_association: association
//     glue between an approval rule template and a repository; no
//     independent enumeration (per-repository association listing only).
