import * as React from "react";
import { Button, StyleSheet, Text, View } from "react-native";
import {
  BottomSheetModal,
  BottomSheetView,
  BottomSheetModalProvider,
  BottomSheetTextInput,
} from "@gorhom/bottom-sheet";
import type { NativeStackScreenProps } from "@react-navigation/native-stack";
import type { RootStackParamList } from "../types";

export default function JoinRoom({navigation}: NativeStackScreenProps<RootStackParamList>) {
  // Bottom sheet stuff
  const bottomSheetModalRef = React.useRef<BottomSheetModal>(null);
  const snapPoints = React.useMemo(() => ["50%", "20%"], []);
  const handlePresentModalPress = React.useCallback(() => {
    bottomSheetModalRef.current?.present();
  }, []);

  const [roomId, setRoomId] = React.useState("");
  const [roomOwner, setRoomOwner] = React.useState("");
  const onInputSubmit = async (e: any) => {
    const input = e.nativeEvent.text;
    const response = await fetch(
      `https://0ba2-165-124-85-89.ngrok-free.app/rooms/${input}`
    );
    if (response.status == 200) {
      setRoomId(input);
      navigation.navigate('Room', {
        owner: roomOwner,
        id: input
      })
    } else {
      setRoomId("");
    }
    const json = await response.json();
    setRoomOwner(json.owner);
  };

  return (
    <BottomSheetModalProvider>
      <Button onPress={handlePresentModalPress} title="Join Room" />
      <BottomSheetModal
        ref={bottomSheetModalRef}
        index={0}
        snapPoints={snapPoints}
        backgroundStyle={styles.drawerContainer}
      >
        <BottomSheetView style={styles.drawerContainer}>
          {roomOwner && <Text>{roomOwner}'s Room</Text>}
          <Text>Enter the room ID</Text>
          <BottomSheetTextInput
            style={styles.input}
            onSubmitEditing={onInputSubmit}
          />
          {roomId && <Text>Found room!</Text>}
        </BottomSheetView>
      </BottomSheetModal>
    </BottomSheetModalProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#fff",
    alignItems: "center",
    justifyContent: "center",
  },
  drawerContainer: {
    flex: 1,
    alignItems: "center",
    backgroundColor: "#eeeeee",
  },
  input: {
    marginTop: 8,
    marginBottom: 10,
    borderRadius: 10,
    fontSize: 16,
    lineHeight: 20,
    padding: 8,
    width: "50%",
    textAlign: "center",
    backgroundColor: "rgba(151, 151, 151, 0.25)",
  },
});
