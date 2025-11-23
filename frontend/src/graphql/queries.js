import { gql } from '@apollo/client';

export const GET_ME = gql`
  query GetMe {
    me {
      id
      username
      email
      mobile
      kycLevel
    }
  }
`;

export const GET_BALANCE = gql`
  query GetBalance {
    balance {
      balance
      bonus
      currency
    }
  }
`;

export const GET_MATCHES = gql`
  query GetMatches {
    matches {
      id
      teamA
      teamB
      oddsA
      oddsB
      oddsDraw
      status
    }
  }
`;

export const GET_MATCH = gql`
  query GetMatch($id: ID!) {
    match(id: $id) {
      id
      teamA
      teamB
      oddsA
      oddsB
      oddsDraw
      status
    }
  }
`;

export const LOGIN_MUTATION = gql`
  mutation Login($email: String!, $password: String!) {
    login(email: $email, password: $password) {
      token
      user {
        id
        username
        email
      }
    }
  }
`;

export const PLACE_BET_MUTATION = gql`
  mutation PlaceBet($matchId: ID!, $selection: String!, $amount: Float!, $odds: Float!) {
    placeBet(matchId: $matchId, selection: $selection, amount: $amount, odds: $odds) {
      id
      status
      message
    }
  }
`;
