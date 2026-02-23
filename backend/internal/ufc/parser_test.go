package ufc

import (
	"strings"
	"testing"
	"time"
)

func TestParseEventLinks(t *testing.T) {
	html := `
<html><body>
  <a href="/event/ufc-fight-night-february-21-2026">UFC FN Feb 21</a>
  <a href="/event/ufc-314">UFC 314</a>
</body></html>`

	links := parseEventLinksHTML(html, "https://www.ufc.com")
	if len(links) != 2 {
		t.Fatalf("expected 2 event links, got %d", len(links))
	}
	if links[0].URL != "https://www.ufc.com/event/ufc-fight-night-february-21-2026" {
		t.Fatalf("unexpected first event url: %s", links[0].URL)
	}
}

func TestParseEventLinks_UsesCardTimestamp(t *testing.T) {
	html := `
<html><body>
  <article class="c-card-event--result">
    <div class="c-card-event--result__info">
      <h3 class="c-card-event--result__headline"><a href="/event/ufc-326">Holloway vs Oliveira 2</a></h3>
      <div class="c-card-event--result__date tz-change-data"
           data-main-card-timestamp="1772935200"
           data-prelims-card-timestamp="1772928000"
           data-early-card-timestamp="1772920800">
        <a href="/event/ufc-326">Sat, Mar 7 / 9:00 PM EST / Main Card</a>
      </div>
    </div>
  </article>
</body></html>`

	links := parseEventLinksHTML(html, "https://www.ufc.com")
	if len(links) != 1 {
		t.Fatalf("expected 1 event link, got %d", len(links))
	}
	if links[0].URL != "https://www.ufc.com/event/ufc-326" {
		t.Fatalf("unexpected event url: %s", links[0].URL)
	}
	expected := time.Unix(1772920800, 0).UTC()
	if !links[0].StartsAt.Equal(expected) {
		t.Fatalf("expected starts_at %s, got %s", expected.Format(time.RFC3339), links[0].StartsAt.Format(time.RFC3339))
	}
	if links[0].Name != "Holloway vs Oliveira 2" {
		t.Fatalf("expected headline as event name, got %q", links[0].Name)
	}
}

func TestParseEventCard(t *testing.T) {
	html := `
<html><head><title>UFC Fight Night: Kape vs. Almabayev | UFC</title></head><body>
  <h1>UFC Fight Night: Kape vs. Almabayev</h1>
  <a href="/athlete/manel-kape">Manel Kape</a>
  <a href="/athlete/asu-almabayev">Asu Almabayev</a>
  <a href="/athlete/cody-brundage">Cody Brundage</a>
  <a href="/athlete/julian-marquez">Julian Marquez</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-fight-night-february-21-2026", "https://www.ufc.com")
	if !strings.Contains(card.Name, "Kape") {
		t.Fatalf("expected parsed event name, got %q", card.Name)
	}
	if len(card.Bouts) != 2 {
		t.Fatalf("expected 2 bouts, got %d", len(card.Bouts))
	}
	if card.Bouts[0].RedURL != "https://www.ufc.com/athlete/manel-kape" {
		t.Fatalf("unexpected first red url: %s", card.Bouts[0].RedURL)
	}
}

func TestParseEventCard_WithAbsoluteAthleteLinks(t *testing.T) {
	html := `
<html><head><title>UFC Fight Night: Kape vs. Almabayev | UFC</title></head><body>
  <h1>UFC Fight Night: Kape vs. Almabayev</h1>
  <a href="https://www.ufc.com/athlete/manel-kape">Manel Kape</a>
  <a href="https://www.ufc.com/athlete/asu-almabayev">Asu Almabayev</a>
  <a href="https://www.ufc.com/athlete/cody-brundage">Cody Brundage</a>
  <a href="https://www.ufc.com/athlete/julian-marquez">Julian Marquez</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-fight-night-february-21-2026", "https://www.ufc.com")
	if len(card.Bouts) != 2 {
		t.Fatalf("expected 2 bouts, got %d", len(card.Bouts))
	}
	if card.Bouts[0].RedURL != "https://www.ufc.com/athlete/manel-kape" {
		t.Fatalf("unexpected first red url: %s", card.Bouts[0].RedURL)
	}
}

func TestParseEventCard_MainCardAndPrelimsWithMeta(t *testing.T) {
	html := `
<html><body>
  <h1>UFC Fight Night: Main vs Prelim</h1>
  <section>
    <h2>Main Card</h2>
    <div>#6</div>
    <a href="/athlete/manel-kape">Manel Kape</a>
    <div>#9</div>
    <a href="/athlete/asu-almabayev">Asu Almabayev</a>
    <div>Flyweight Bout</div>
  </section>
  <section>
    <h2>Prelims</h2>
    <div>#12</div>
    <a href="/athlete/cody-brundage">Cody Brundage</a>
    <div>#14</div>
    <a href="/athlete/julian-marquez">Julian Marquez</a>
    <div>Middleweight Bout</div>
  </section>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-fight-night-february-21-2026", "https://www.ufc.com")
	if len(card.Bouts) != 2 {
		t.Fatalf("expected 2 bouts, got %d", len(card.Bouts))
	}
	if card.Bouts[0].CardSegment != "main_card" {
		t.Fatalf("expected first bout to be main_card, got %q", card.Bouts[0].CardSegment)
	}
	if card.Bouts[1].CardSegment != "prelims" {
		t.Fatalf("expected second bout to be prelims, got %q", card.Bouts[1].CardSegment)
	}
	if card.Bouts[0].WeightClass != "Flyweight" {
		t.Fatalf("expected first bout weight class Flyweight, got %q", card.Bouts[0].WeightClass)
	}
	if card.Bouts[1].WeightClass != "Middleweight" {
		t.Fatalf("expected second bout weight class Middleweight, got %q", card.Bouts[1].WeightClass)
	}
	if card.Bouts[0].RedRank != "#6" || card.Bouts[0].BlueRank != "#9" {
		t.Fatalf("unexpected first bout ranks: red=%q blue=%q", card.Bouts[0].RedRank, card.Bouts[0].BlueRank)
	}
	if card.Bouts[1].RedRank != "#12" || card.Bouts[1].BlueRank != "#14" {
		t.Fatalf("unexpected second bout ranks: red=%q blue=%q", card.Bouts[1].RedRank, card.Bouts[1].BlueRank)
	}
}

func TestParseEventCard_ExtractsPosterFromHeroImage(t *testing.T) {
	html := `
<html><body>
  <div class="c-hero__image">
    <img src="https://ufc.com/images/styles/background_image_sm/s3/2026-02/022126-event-art.jpg?h=d1cb525d&amp;itok=abc123" />
  </div>
  <a href="/athlete/manel-kape">Manel Kape</a>
  <a href="/athlete/asu-almabayev">Asu Almabayev</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-fight-night-february-21-2026", "https://www.ufc.com")
	if card.PosterURL == "" {
		t.Fatalf("expected poster url from hero image, got empty")
	}
	if !strings.Contains(card.PosterURL, "/images/styles/background_image_sm/") {
		t.Fatalf("unexpected poster url: %s", card.PosterURL)
	}
	if !strings.Contains(card.PosterURL, "&itok=abc123") {
		t.Fatalf("expected html entity to be unescaped in poster url: %s", card.PosterURL)
	}
}

func TestParseEventCard_ExtractsStartsAtFromFightCardTime(t *testing.T) {
	html := `
<html><body>
  <div class="field field--name-fight-card-time-main"><time datetime="2026-02-21T20:00:00Z">Sat, Feb 21 / 8:00 PM</time></div>
  <div class="field field--name-fight-card-time-prelims"><time datetime="2026-02-21T17:00:00Z">Sat, Feb 21 / 5:00 PM</time></div>
  <a href="/athlete/fighter-a">Fighter A</a>
  <a href="/athlete/fighter-b">Fighter B</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-326", "https://www.ufc.com")
	expected := time.Date(2026, 2, 21, 20, 0, 0, 0, time.UTC)
	if !card.StartsAt.Equal(expected) {
		t.Fatalf("expected starts at %s, got %s", expected.Format(time.RFC3339), card.StartsAt.Format(time.RFC3339))
	}
}

func TestParseEventCard_StatusUsesEventLiveStatsFinal(t *testing.T) {
	html := `
<html><body>
  <script type="application/json">{"eventLiveStats":{"event_fmid":"1297","final":true}}</script>
  <div class="field field--name-fight-card-time-main"><time datetime="2026-02-21T20:00:00Z">Sat, Feb 21 / 8:00 PM</time></div>
  <a href="/athlete/fighter-a">Fighter A</a>
  <a href="/athlete/fighter-b">Fighter B</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-326", "https://www.ufc.com")
	if card.Status != "completed" {
		t.Fatalf("expected completed from eventLiveStats.final=true, got %q", card.Status)
	}
}

func TestParseEventCard_StatusScheduledWhenEventLiveStatsNotFinal(t *testing.T) {
	html := `
<html><body>
  <script type="application/json">{"eventLiveStats":{"event_fmid":"1297","final":false}}</script>
  <div class="field field--name-fight-card-time-main"><time datetime="2099-02-21T20:00:00Z">Sat, Feb 21 / 8:00 PM</time></div>
  <a href="/athlete/fighter-a">Fighter A</a>
  <a href="/athlete/fighter-b">Fighter B</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-326", "https://www.ufc.com")
	if card.Status != "scheduled" {
		t.Fatalf("expected scheduled from eventLiveStats.final=false, got %q", card.Status)
	}
}

func TestParseEventCard_StatusLiveWhenStartedAndNotFinal(t *testing.T) {
	startedAt := time.Now().UTC().Add(-20 * time.Minute).Format(time.RFC3339)
	html := `
<html><body>
  <script type="application/json">{"eventLiveStats":{"event_fmid":"1297","final":false}}</script>
  <div class="field field--name-fight-card-time-main"><time datetime="` + startedAt + `">Live card</time></div>
  <a href="/athlete/fighter-a">Fighter A</a>
  <a href="/athlete/fighter-b">Fighter B</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-326", "https://www.ufc.com")
	if card.Status != "live" {
		t.Fatalf("expected live when started and not final, got %q", card.Status)
	}
}

func TestParseEventCard_StatusCompletedWhenStartedLongAgoAndNotFinal(t *testing.T) {
	startedAt := time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	html := `
<html><body>
  <script type="application/json">{"eventLiveStats":{"event_fmid":"1297","final":false}}</script>
  <div class="field field--name-fight-card-time-main"><time datetime="` + startedAt + `">Old card</time></div>
  <a href="/athlete/fighter-a">Fighter A</a>
  <a href="/athlete/fighter-b">Fighter B</a>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-326", "https://www.ufc.com")
	if card.Status != "completed" {
		t.Fatalf("expected completed when started long ago and not final, got %q", card.Status)
	}
}

func TestParseEventCard_ExtractsBoutResultMeta(t *testing.T) {
	html := `
<html><body>
  <h3>Main Card</h3>
  <div class="c-listing-fight">
    <div class="c-listing-fight__corner-body--red">
      <div class="c-listing-fight__outcome-wrapper"><div class="c-listing-fight__outcome--win">Win</div></div>
    </div>
    <div class="c-listing-fight__corner-body--blue">
      <div class="c-listing-fight__outcome-wrapper"><div class="c-listing-fight__outcome--loss">Loss</div></div>
    </div>
    <div class="c-listing-fight__result-text round">2</div>
    <div class="c-listing-fight__result-text time">1:40</div>
    <div class="c-listing-fight__result-text method">KO/TKO</div>
    <a href="/athlete/fighter-a">Fighter A</a>
    <a href="/athlete/fighter-b">Fighter B</a>
  </div>
</body></html>`

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-326", "https://www.ufc.com")
	if len(card.Bouts) != 1 {
		t.Fatalf("expected 1 bout, got %d", len(card.Bouts))
	}
	bout := card.Bouts[0]
	if bout.WinnerSide != "red" {
		t.Fatalf("expected winner side red, got %q", bout.WinnerSide)
	}
	if bout.Method != "KO/TKO" {
		t.Fatalf("expected method KO/TKO, got %q", bout.Method)
	}
	if bout.Round != 2 {
		t.Fatalf("expected round 2, got %d", bout.Round)
	}
	if bout.TimeSec != 100 {
		t.Fatalf("expected time_sec 100, got %d", bout.TimeSec)
	}
	if bout.Result == "" {
		t.Fatalf("expected non-empty result")
	}
}

func TestParseEventCard_ExtractsBlueWinnerFromLossMarker(t *testing.T) {
	html := `
<html><body>
  <h3>Main Card</h3>
  <div class="c-listing-fight">
    <div class="c-listing-fight__corner--red">
      <div class="c-listing-fight__outcome-wrapper"><div class="c-listing-fight__outcome--loss">Loss</div></div>
    </div>
    <div class="c-listing-fight__corner--blue">
      <div class="c-listing-fight__outcome-wrapper"><div class="c-listing-fight__outcome--win">Win</div></div>
    </div>
    <div class="c-listing-fight__result-text round">3</div>
    <div class="c-listing-fight__result-text time">5:00</div>
    <div class="c-listing-fight__result-text method">Decision - Unanimous</div>
    <a href="/athlete/fighter-a">Fighter A</a>
    <a href="/athlete/fighter-b">Fighter B</a>
  </div>
</body></html>`

	boutResults := parseBoutResultMeta(html)
	if len(boutResults) != 1 {
		t.Fatalf("expected 1 parsed bout result, got %d", len(boutResults))
	}
	if boutResults[0].WinnerSide != "blue" {
		t.Fatalf("expected parsed winner side blue, got %q", boutResults[0].WinnerSide)
	}

	card := parseEventCardHTML(html, "https://www.ufc.com/event/ufc-326", "https://www.ufc.com")
	if len(card.Bouts) != 1 {
		t.Fatalf("expected 1 bout, got %d", len(card.Bouts))
	}
	if card.Bouts[0].WinnerSide != "blue" {
		t.Fatalf("expected blue winner side, got %q", card.Bouts[0].WinnerSide)
	}
}

func TestParseAthleteProfile(t *testing.T) {
	html := `
<html><head>
  <meta property="og:image" content="https://cdn.ufc.com/fighter/zachary-reese.jpg" />
  <title>Zachary Reese | UFC</title>
</head><body>
  <h1>Zachary Reese</h1>
  <div>Fighting Out Of</div><div>USA</div>
  <div>Record:</div><div>8-2-0</div>
</body></html>`

	profile := parseAthleteProfileHTML(html, "https://www.ufc.com/athlete/zachary-reese")
	if profile.Name != "Zachary Reese" {
		t.Fatalf("unexpected name: %q", profile.Name)
	}
	if profile.Country != "USA" {
		t.Fatalf("unexpected country: %q", profile.Country)
	}
	if profile.Record != "8-2-0" {
		t.Fatalf("unexpected record: %q", profile.Record)
	}
}

func TestParseAthleteLinks(t *testing.T) {
	html := `
<html><body>
  <a href="/athlete/zachary-reese">Zachary Reese</a>
  <a href="/athlete/ramazan-temirov">Ramazan Temirov</a>
</body></html>`

	links := parseAthleteLinksHTML(html, "https://www.ufc.com")
	if len(links) != 2 {
		t.Fatalf("expected 2 athlete links, got %d", len(links))
	}
	if links[1] != "https://www.ufc.com/athlete/ramazan-temirov" {
		t.Fatalf("unexpected second athlete url: %s", links[1])
	}
}

func TestParseAthleteProfile_ExtractsStatsAndRecords(t *testing.T) {
	html := `
<html><body>
  <h1>Joshua Van</h1>
  <p class="hero-profile__nickname">"The Fearless"</p>
  <div class="hero-profile__division-title">Flyweight Division</div>
  <div class="hero-profile__division-body">16-2-0 (W-L-D)</div>

  <div class="hero-profile__stat">
    <p class="hero-profile__stat-numb">8</p>
    <p class="hero-profile__stat-text">Wins by Knockout</p>
  </div>
  <div class="hero-profile__stat">
    <p class="hero-profile__stat-numb">2</p>
    <p class="hero-profile__stat-text">Wins by Submission</p>
  </div>

  <div class="c-stat-compare__number">8.84</div>
  <div class="c-stat-compare__label">Sig. Str. Landed</div>
  <div class="c-stat-compare__number">0.84</div>
  <div class="c-stat-compare__label">Takedown avg</div>

  <div class="c-bio__label">Place of Birth</div>
  <div class="c-bio__text">Hakha , Myanmar</div>
</body></html>`

	profile := parseAthleteProfileHTML(html, "https://www.ufc.com/athlete/joshua-van")
	if profile.Nickname != "The Fearless" {
		t.Fatalf("unexpected nickname: %q", profile.Nickname)
	}
	if profile.WeightClass != "Flyweight" {
		t.Fatalf("unexpected weight class: %q", profile.WeightClass)
	}
	if profile.Record != "16-2-0" {
		t.Fatalf("unexpected record: %q", profile.Record)
	}
	if profile.Country != "Myanmar" {
		t.Fatalf("unexpected country: %q", profile.Country)
	}
	if profile.Records["Wins by Knockout"] != "8" {
		t.Fatalf("expected wins by knockout in records, got %+v", profile.Records)
	}
	if profile.Stats["Sig. Str. Landed"] != "8.84" {
		t.Fatalf("expected sig. str. landed in stats, got %+v", profile.Stats)
	}
}
