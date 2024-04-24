import * as React from "react";
import * as Linking from "expo-linking";
import { Button, StyleSheet, Text, View, Share } from "react-native";
import * as AuthSession from "expo-auth-session";
import Pocketbase from "pocketbase";
import uuid from "react-native-uuid";
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { RootStackParamList } from "../types";


const pb = new Pocketbase("https://0ba2-165-124-85-89.ngrok-free.app/");
const prefix = Linking.createURL("/");

export default function Home({ navigation }: NativeStackScreenProps<RootStackParamList, 'Home'>) {
  const [userId, setUserId] = React.useState("");
  const [username, setUsername] = React.useState("");
  const [accessToken, setAccessToken] = React.useState("");
  const [roomId, setRoomId] = React.useState("");

  const discovery = {
    authorizationEndpoint: "https://accounts.spotify.com/authorize",
  };

  const [request, response, promptAsync] = AuthSession.useAuthRequest(
    {
      clientId: clientId,
      scopes: [
        "user-read-private",
        "user-read-email",
        "user-modify-playback-state",
        "user-top-read"
      ],
      redirectUri: AuthSession.makeRedirectUri({
        path: "redirect",
      }),
      usePKCE: false
    },
    discovery
  );

  React.useEffect(() => {
    if (response?.type === "success") {
      const { code, state } = response.params;
      if (!accessToken) {
        getAccessToken(code, '');
      }
    }
  }, [response]);

  const getAccessToken = async (code: string, state: string) => {
    if (!request) {
      return;
    }

    const codeVerifier = request.codeVerifier as string;
    const redirectUri = AuthSession.makeRedirectUri({
      path: "redirect",
    });

    const body = JSON.stringify({
      client_id: clientId,
      grant_type: "authorization_code",
      code,
      // code_verifier: codeVerifier,
      state: state,
      redirect_uri: redirectUri,
    });

    const payload = {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: body,
    };
    try {
      // await pb.collection('users').authWithOAuth2Code('spotify', code, codeVerifier, redirectUri)
      const response = await fetch(
        "https://0ba2-165-124-85-89.ngrok-free.app/api/login",
        payload
      );
      const json = await response.json();
      setUserId(json.userId);
      setUsername(json.display_name);
      setAccessToken(json.access_token);
    } catch (err) {
      console.log(err);
    }
  };

  const createRoom = async () => {
    if (!accessToken) {
      return;
    }
    try {
      const room = uuid.v4() as string;
      const response = await fetch(
        `https://0ba2-165-124-85-89.ngrok-free.app/rooms/${room}`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            user_id: userId,
          }),
        }
      );
      setRoomId(room);
    } catch (err) {
      console.log(err);
    }
  };

  const shareRoom = async () => {
    try {
      if (roomId && username) {
        const result = await Share.share({
          message: `${prefix}rooms/${roomId}/${username}`,
        });
      }
    } catch (err) {
      console.error(err);
    }
  };

  const closeRoom = async () => {
    if (!roomId) {
      return;
    }

    const response = await fetch(
      `https://0ba2-165-124-85-89.ngrok-free.app/rooms/${roomId}`,
      {
        method: "DELETE",
      }
    );
    if (response.status == 200) {
      setRoomId("");
    }
  };

  return (
    <View style={styles.container}>
      {username && <Text>Hi, {username}</Text>}
      <Button
        onPress={() => {
          promptAsync();
        }}
        title="Log in with Spotify"
      ></Button>
      {roomId && (
        <>
          <Button
            onPress={() => {
              closeRoom();
            }}
            title="Close room"
          ></Button>
          <Button
            onPress={() => {
              shareRoom();
            }}
            title="Share room"
          ></Button>
          {roomId && (
            <>
              <Text>{roomId}</Text>
            </>
          )}
        </>
      )}
      {userId && (
        <Button
          onPress={() => {
            createRoom();
          }}
          title="Create room"
        ></Button>
      )}
      {/* <JoinRoom navigation={navigation} /> */}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    alignItems: "center",
    justifyContent: "center",
  },
  contentContainer: {
    flex: 1,
    alignItems: "center",
  },
  containerHeadline: {
    fontSize: 24,
    fontWeight: "600",
    padding: 20,
  },
});
