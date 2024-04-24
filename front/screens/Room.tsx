import * as React from "react";
import {
  View,
  Text,
  TextInput,
  StyleSheet,
  Pressable,
  Image,
} from "react-native";
import TrackResult from "../components/TrackResult";
import { ScrollView } from "react-native-gesture-handler";
import { Feather, EvilIcons } from "@expo/vector-icons";
import type { NativeStackScreenProps } from "@react-navigation/native-stack";
import type { RootStackParamList } from "../types";

import {
  BottomSheetModal,
  BottomSheetView,
  BottomSheetModalProvider,
} from "@gorhom/bottom-sheet";
import AddedSongToQueue from "../components/AddedSongToQueue";

export default function Room({
  navigation,
  route,
}: NativeStackScreenProps<RootStackParamList, "Room">) {
  const [searchResults, setSearchResults] = React.useState([]);
  const searchInputRef = React.useRef<any>();
  const [track, setTrack] = React.useState(null);
  const [topTracks, setTopTracks] = React.useState([]);
  const [currentInput, setCurrentInput] = React.useState("");

  const queueAddedBottomSheetRef = React.useRef<BottomSheetModal>(null);
  const snapPoints = React.useMemo(() => ["35%"], []);
  const handlePresentModalPress = React.useCallback(() => {
    queueAddedBottomSheetRef.current?.present();
  }, []);

  const [activePlayerExists, setActivePlayerExists] = React.useState(true);

  React.useEffect(() => {
    (async function effect() {
      const joinRoomResponse = await fetch(
        `https://0ba2-165-124-85-89.ngrok-free.app/rooms/${route.params.id}`
      );
      if (joinRoomResponse.status == 404) {
        navigation.reset({
          index: 0,
          routes: [{ name: "Home" }],
        });
      }

      const topTracksResponse = await fetch(
        `https://0ba2-165-124-85-89.ngrok-free.app/rooms/${route.params.id}/top`
      );
      const topTracksJson = await topTracksResponse.json();
      setTopTracks(topTracksJson.items);
    })();
  }, []);


  const addToQueue = async (track: any) => {
    try {
      if (!route.params.id || !track) {
        return;
      }
      const response = await fetch(
        `https://0ba2-165-124-85-89.ngrok-free.app/rooms/${route.params.id}/queue`,
        {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            spotify_uri: track.uri,
          }),
        }
      );
      console.log(response.status);
      if (response.status == 404) {
        setActivePlayerExists(false);
        setTrack(null);
      } else {
        setTrack(track);
        setActivePlayerExists(true);
      }
      handlePresentModalPress();
    } catch (err) {
      console.log(err);
    }
  };

  const clearInput = () => {
    searchInputRef.current.clear();
    setCurrentInput("");
    setSearchResults([]);
  };

  const onSubmit = async (e: any) => {
    const input = e.nativeEvent.text;
    if (!input) {
      return;
    }
    const response = await fetch(
      `https://0ba2-165-124-85-89.ngrok-free.app/rooms/${route.params.id}/search?track=${input}`
    );
    const json = await response.json();
    setSearchResults(json.tracks.items);
  };

  return (
    <View style={styles.container}>
      {/* <Text>{route.params.id}</Text>
      <Text>{route.params.owner}'s Room</Text> */}
      <View style={{ width: "100%" }}>
        <Text style={{ fontSize: 24, fontWeight: "600", color: "#555555" }}>
          Add a song to the queue!
        </Text>
        <View style={styles.searchbar}>
          <EvilIcons name="search" size={24} color="#888888" />
          <TextInput
            ref={searchInputRef}
            style={{
              fontSize: 16,
              width: "80%",
            }}
            placeholder="Search for a song"
            onSubmitEditing={onSubmit}
            onChangeText={(text) => {
              setCurrentInput(text);
            }}
          ></TextInput>
          <Pressable
            style={{ marginHorizontal: 2 }}
            onPress={() => {
              clearInput();
            }}
          >
            <Feather name="x" size={18} color="#888888" />
          </Pressable>
        </View>
      </View>

      {!currentInput && (
        <View style={styles.results}>
          <Text style={{ fontSize: 18, fontWeight: "600", marginVertical: 4 }}>
            Suggestions
          </Text>
          <ScrollView style={styles.scrollView}>
            {topTracks.map((track: any, i: number) => {
              return (
                <TrackResult key={i} track={track} addToQueue={addToQueue} />
              );
            })}
          </ScrollView>
        </View>
      )}
      {currentInput && searchResults.length > 0 && (
        <View style={styles.results}>
          <Text style={{ fontSize: 24, fontWeight: "600", marginVertical: 4 }}>
            Search results
          </Text>
          <ScrollView style={styles.scrollView}>
            {searchResults.map((track: any, i: number) => {
              return (
                <TrackResult key={i} track={track} addToQueue={addToQueue} />
              );
            })}
          </ScrollView>
        </View>
      )}
      <BottomSheetModalProvider>
          <BottomSheetModal
            ref={queueAddedBottomSheetRef}
            index={0}
            snapPoints={snapPoints}
            backgroundStyle={styles.drawerContainer}
          >
            <BottomSheetView style={styles.drawerContainer}>
              {track ? (
                <AddedSongToQueue track={track} />
              ) : (
                <View style={{ marginVertical: 10, marginHorizontal: 20 }}>
                  <Text style={{ fontSize: 24, fontWeight: "600", marginVertical: 10 }}>
                    No active player found!
                  </Text>
                  <Text style={{ color: "#555555" }}>
                    Make sure you have an open spotify session to add a song to
                    the queue.
                  </Text>
                </View>
              )}
            </BottomSheetView>
          </BottomSheetModal>
      </BottomSheetModalProvider>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: "center",
    margin: 10,
    paddingBottom: 100,
  },
  scrollView: {
    marginBottom: 20,
  },
  searchbar: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    width: "100%",
    borderRadius: 5,
    paddingHorizontal: 4,
    paddingVertical: 6,
    marginVertical: 10,
    backgroundColor: "rgba(151, 151, 151, 0.25)",
  },
  results: {
    width: "100%",
    marginBottom: 10,
  },
  drawerContainer: {
    flex: 1,
    alignItems: "center",
    backgroundColor: "#ffffff",
  },
});
