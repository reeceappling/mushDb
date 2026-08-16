import {Note} from "@/app/components/formSubcomponents/notes";
import {Liquid} from "@/app/components/formSubcomponents/liquids";
import {Nutrient} from "@/app/components/formSubcomponents/nutrients";
import {Sugar} from "@/app/components/formSubcomponents/sugars";
import {Antibiotic} from "@/app/components/formSubcomponents/antibiotic";
import {Additive} from "@/app/components/formSubcomponents/additives";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {AgarRecipeSelector, NewAgarRecipeForm} from "@/app/components/agarRecipeClient";
import {CapitalizeFirstLetter} from "@/app/components/commonServer";

export function TestAgarRecipeOk(){
    return new AgarRecipeData({
        _id: "(AGAR RECIPE ID HERE)",
        name: "(AGAR RECIPE NAME HERE)",
        liquids: [{name:"Water",pct:100}],
        agar: 20, // g/L
        standard: true,
        nutrients: [{nutrient:"Oats",amount:17.2,unit:"handfuls"}],
        sugars: [{type:"Glucose",amount:100,unit:"tons"}],
        additives: [],// TODO: this
        antibiotics: [], // TODO: this
        notes: [{time: Date.now(),note: "(TEST NOTE 1)"},{time: Date.now()+2000,note: "(TEST NOTE 2)"}],
        lastUpdated: 789,
        acl: TestAcl(),
    })
}

export interface AgarRecipeData {
    _id: string
    name: string
    liquids: Liquid[]
    agar: number
    standard: boolean
    nutrients?: Nutrient[]
    sugars?: Sugar[]
    additives?: Additive[]
    antibiotics?: Antibiotic[]
    notes?: Note[]
    lastUpdated: number
    acl: ACL
}
export class AgarRecipeData {
    // Accept a single object containing the fields
    constructor(init?: Partial<AgarRecipeData>) {
        // Dynamically map the object fields onto the class instance
        Object.assign(this, init);
    }

    public getId(): string {
        return this._id
    }
    public getIdUrlEncoded(): string {
        return encodeURI(this.getId())
    }
    public entryType(): string {
        return "agarRecipe"
    }
    public description(): string {
        const ultraSoftCutoff = 15 // TODO; ensure ok
        const softCutoff = 19 // TODO; ensure ok
        const hardCutoff = 22 // TODO; ensure ok
        const ultraHardCutoff = 24 // TODO; ensure ok
        const isRegularSoftness = this.agar>=softCutoff && this.agar <hardCutoff
        const isAntibiotic = this.antibiotics===undefined||this.antibiotics.length==0
        const standardPart = this.standard?`standard `:''
        const softnessPart = isRegularSoftness?``:(this.agar<softCutoff?(this.agar<ultraSoftCutoff?`ultrasoft `:`soft `):(this.agar<ultraHardCutoff?"hard ":"ultrahard "))
        const antibioticsPart = (isAntibiotic)?"":"antibiotic "
        const firstSentence = CapitalizeFirstLetter(`${standardPart}${softnessPart}${antibioticsPart}agar recipe ${this.name}`)
        const nuteSent = `${(this.nutrients===undefined||this.nutrients.length==0)?`no`:this.nutrients.length} nutrients`
        const liqSent = (this.liquids.length==1)?`${this.liquids[0].name} based`:`${this.liquids.length} liquids`
        const sugSent = `${(this.sugars===undefined||this.sugars.length==0)?`no`:this.sugars.length} sugars`
        const addSent = `${(this.additives===undefined||this.additives.length==0)?`no`:this.additives.length} additives`
        const lastSent = `Last updated on ${new Date(this.lastUpdated).toISOString()}`
        return `${firstSentence}. ${nuteSent}. ${liqSent}. ${sugSent}. ${addSent}. ${lastSent}`
    }
}

export function AgarRecipeSelectorCloseable(sp: SelectorProps<AgarRecipeData>) {
    const doSel = (val?: AgarRecipeData): void => {
        val && sp.doSelect(val)
    }
    return <CloseableSelector<AgarRecipeData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Recipe List",
        createTxt: "Create Agar Recipe",
        lowercase: "agar recipe",
        creatorInPage: sp.creatorInPage,
        createEndpt: "agarRecipe",
        createSelector: (selHdl: (onSelect: AgarRecipeData) => void) => {
            return <AgarRecipeSelector allowCreate={sp.allowCreation} doSelect={(v) => {
                v && selHdl(v)
            }}/>
        },
        createCreator: (selHdl: (onSelect: AgarRecipeData) => void) => {
            return <NewAgarRecipeForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        },
    }}/>
}
