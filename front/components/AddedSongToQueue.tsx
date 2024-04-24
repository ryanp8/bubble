import { View, Text, Image, StyleSheet } from "react-native";
export default function AddedSongToQueue({ track }) {
  return (
    <>
      <Text>Added!</Text>
      <View
        style={{
          alignItems: "center",
          marginVertical: 10,
        }}
      >
        <Image
          style={styles.logo}
          source={{
            uri: track.album.images[0].url,
          }}
        />
        <View style={{ marginVertical: 10 }}>
          <Text
            style={{
              fontSize: 18,
              fontWeight: "500",
              color: "#333333",
            }}
          >
            {track.name.length > 25
              ? track.name.slice(0, 25) + "..."
              : track.name}
          </Text>
          <Text style={{ color: "#555555", textAlign: "center" }}>
            {track.artists[0].name}
          </Text>
        </View>
      </View>
    </>
  );
}

const styles = StyleSheet.create({
  logo: {
    width: 100,
    height: 100,
  },
});
