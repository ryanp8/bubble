import * as React from "react";
import { View, Text, Image, StyleSheet } from "react-native";
import { TouchableHighlight } from "react-native-gesture-handler";

export default function TrackResult({ track, addToQueue }) {
  return (
    <TouchableHighlight
      underlayColor="#DDDDDD"
      onPress={() => {
        addToQueue(track);
      }}
    >
      <View style={styles.container}>
        <Image
          style={styles.logo}
          source={{
            uri: track.album.images[0].url,
          }}
        />
        <View style={{ marginLeft: 4 }}>
          <Text style={{ fontSize: 18, fontWeight: '500', color: '#333333'}}>
            {track.name.length > 25
              ? track.name.slice(0, 25) + "..."
              : track.name}
          </Text>
          <Text style={{ color: "#555555" }}>{track.artists[0].name}</Text>
        </View>

        {/* <Text>{track.uri}</Text> */}
      </View>
    </TouchableHighlight>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    flexDirection: "row",
    marginHorizontal: 5,
    marginVertical: 8,
  },
  logo: {
    width: 50,
    height: 50,
  },
});
