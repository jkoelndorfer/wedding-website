#!/usr/bin/env python3

import sys
from typing import Self

import boto3


dynamodb = boto3.client("dynamodb")


class Person:
    """
    Model representing an invited person.
    """
    def __init__(self, id: int, name: str) -> None:
        self.id = id
        self.name = name


class Invitation:
    """
    Model representing an invitation.
    """
    def __init__(
        self,
        id: str,
        people: list[Person],
    ) -> None:
        pass

    @classmethod
    def from_data(cls, data: dict) -> Self:
        """
        Hydrates an Invitation.
        """


class InvitationResponse:
    """
    Model representing an invitation response.
    """
    def __init__(self, invite: Invitation) -> None:
        pass


class InvitationRepository:
    """
    Repository providing access to invitations.
    """
    def __init__(self, dynamodb_table_arn: str) -> None:
        """
        Initializes an InvitationRepository.
        """
        self.dynamodb_table_arn = dynamodb_table_arn

    def _get_item(self, key: dict, attributes_to_get: dict) -> dict:
        """
        Gets an item from DynamoDB.
        """
        dynamodb.get_item(
            TableName=self.dynamodb_table_arn,
            Key=key,
            AttributesToGet=attributes_to_get,
            ConsistentRead=True,
        )

    def get_invitation(self, id: str) -> Invitation:
        """
        Looks up the invitation with the given id.
        """
        # Invitations have the following structure:
        #
        # {
        #     InviteId: string  # users enter this to RSVP
        #     Email:    string  # the email address that the invitation was sent to
        #     People:   list of ... {
        #         PersonId: int     # used to identify the person during form submission
        #         Name:     string  # the person's name
        #     }
        #     CeremonyInvite: bool  # whether the people in this invite may attend the ceremony
        #     Response:       null if no response, else a list of ... {
        #         PersonId: int             # the ID of the person the response applies to
        #         AttendingCeremony:  bool  # whether this person plans to attend the ceremony
        #         AttendingReception: bool  # whether this person plans to attend the reception
        #     }
        # }
        data = self._get_item({"InviteId": {"S": id}})


def lambda_handler(event, context):
    """
    This function is the Lambda event handler.

    It is used during normal operation when this code runs in AWS Lambda.
    """

def main(argv):
    """
    This function is the handler for invocation from the command line.

    It is used for local development only.
    """


if __name__ == "__main__":
    main(sys.argv[1:])
