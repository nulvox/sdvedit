package save

import (
	"bytes"
	"strings"
	"testing"
)

const friendshipFixture = `<?xml version="1.0" encoding="utf-8"?>
<SaveGame xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema">
  <player>
    <name>Trent</name>
    <money>500</money>
    <deepestMineLevel>5</deepestMineLevel>
    <friendshipData>
      <item>
        <key><string>Lewis</string></key>
        <value>
          <Friendship>
            <Points>16</Points>
            <GiftsThisWeek>0</GiftsThisWeek>
            <GiftsToday>0</GiftsToday>
            <TalkedToToday>false</TalkedToToday>
            <Status>Friendly</Status>
          </Friendship>
        </value>
      </item>
      <item>
        <key><string>Robin</string></key>
        <value>
          <Friendship>
            <Points>250</Points>
            <GiftsThisWeek>1</GiftsThisWeek>
            <GiftsToday>0</GiftsToday>
            <TalkedToToday>true</TalkedToToday>
            <Status>Friendly</Status>
          </Friendship>
        </value>
      </item>
    </friendshipData>
  </player>
  <currentSeason>spring</currentSeason>
  <dayOfMonth>16</dayOfMonth>
  <year>1</year>
  <dailyLuck>-0.094</dailyLuck>
  <weatherForTomorrow>Sun</weatherForTomorrow>
  <mine_lowestLevelReached>5</mine_lowestLevelReached>
</SaveGame>`

const nilAttrFixture = `<?xml version="1.0" encoding="utf-8"?>
<root>
  <skinId><string xsi:nil="true"/></skinId>
  <Quest xsi:type="SocializeQuest"><id>9</id></Quest>
</root>`

func TestParse_BasicStructure(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if root.Name != "SaveGame" {
		t.Errorf("root name = %q, want SaveGame", root.Name)
	}
}

func TestParse_NestedGet(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}

	// deep path traversal
	node := root.Get("player/name")
	if node == nil {
		t.Fatal("player/name not found")
	}
	if node.Text != "Trent" {
		t.Errorf("player/name = %q, want Trent", node.Text)
	}
}

func TestParse_LeafValues(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		path string
		want string
	}{
		{"player/money", "500"},
		{"player/deepestMineLevel", "5"},
		{"currentSeason", "spring"},
		{"dayOfMonth", "16"},
		{"year", "1"},
		{"dailyLuck", "-0.094"},
		{"weatherForTomorrow", "Sun"},
		{"mine_lowestLevelReached", "5"},
	}
	for _, tc := range cases {
		n := root.Get(tc.path)
		if n == nil {
			t.Errorf("path %q not found", tc.path)
			continue
		}
		if n.Text != tc.want {
			t.Errorf("path %q = %q, want %q", tc.path, n.Text, tc.want)
		}
	}
}

func TestParse_AttributePreservation(t *testing.T) {
	root, err := Parse(strings.NewReader(nilAttrFixture))
	if err != nil {
		t.Fatal(err)
	}

	questNode := root.Child("Quest")
	if questNode == nil {
		t.Fatal("Quest node not found")
	}
	xsiType := questNode.Attr("type")
	if xsiType != "SocializeQuest" {
		t.Errorf("xsi:type = %q, want SocializeQuest", xsiType)
	}
}

func TestParse_FriendshipItems(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}

	fd := root.Get("player/friendshipData")
	if fd == nil {
		t.Fatal("friendshipData not found")
	}
	items := fd.ChildrenNamed("item")
	if len(items) != 2 {
		t.Fatalf("friendshipData items = %d, want 2", len(items))
	}

	// first item should be Lewis with 16 points
	lewisKey := items[0].Get("key/string")
	if lewisKey == nil || lewisKey.Text != "Lewis" {
		t.Errorf("first item key = %v, want Lewis", lewisKey)
	}
	lewisPoints := items[0].Get("value/Friendship/Points")
	if lewisPoints == nil || lewisPoints.Text != "16" {
		t.Errorf("Lewis points = %v, want 16", lewisPoints)
	}
}

func TestSetText(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}

	if err := root.SetText("player/money", "99999"); err != nil {
		t.Fatalf("SetText error: %v", err)
	}
	node := root.Get("player/money")
	if node.Text != "99999" {
		t.Errorf("after SetText, money = %q, want 99999", node.Text)
	}
}

func TestRoundTrip(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	if err := Serialize(root, &buf); err != nil {
		t.Fatalf("Serialize error: %v", err)
	}

	// re-parse the output and verify key values survive
	root2, err := Parse(&buf)
	if err != nil {
		t.Fatalf("re-parse error: %v", err)
	}

	cases := []struct {
		path string
		want string
	}{
		{"player/name", "Trent"},
		{"player/money", "500"},
		{"currentSeason", "spring"},
		{"year", "1"},
	}
	for _, tc := range cases {
		n := root2.Get(tc.path)
		if n == nil {
			t.Errorf("round-trip: path %q not found", tc.path)
			continue
		}
		if n.Text != tc.want {
			t.Errorf("round-trip: path %q = %q, want %q", tc.path, n.Text, tc.want)
		}
	}
}

func TestRoundTrip_MutationPreserved(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}

	root.SetText("player/money", "12345")
	root.SetText("currentSeason", "summer")

	var buf bytes.Buffer
	Serialize(root, &buf)

	root2, err := Parse(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if n := root2.Get("player/money"); n == nil || n.Text != "12345" {
		t.Errorf("mutated money not preserved in round-trip")
	}
	if n := root2.Get("currentSeason"); n == nil || n.Text != "summer" {
		t.Errorf("mutated season not preserved in round-trip")
	}
}

func TestChildrenNamed(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}
	items := root.Get("player/friendshipData").ChildrenNamed("item")
	if len(items) != 2 {
		t.Errorf("ChildrenNamed = %d, want 2", len(items))
	}
}

func TestGetMissingPath(t *testing.T) {
	root, err := Parse(strings.NewReader(friendshipFixture))
	if err != nil {
		t.Fatal(err)
	}
	if n := root.Get("player/nonexistent/deep"); n != nil {
		t.Errorf("expected nil for missing path, got %v", n)
	}
}
