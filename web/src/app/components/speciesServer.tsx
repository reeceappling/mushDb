import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {ACL, TestAcl} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {SlantSelector} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {SpeciesSelector} from "@/app/components/speciesClient";

export function TestSpeciesOk() {
    return new SpeciesData({
        _id: "(ID_HERE)",
        scientificName: "(SCI_NAME_HERE)",
        aliases: ["(Alias 1)", "(Alias 2)"],
        standardSubstrate: "(SUBSTRATE ID)",
        notes: [{
            time: 123,
            note: "(NOTE 1)"
        }, {
            time: 456,
            note: "(NOTE 2)"
        }],
        lastUpdated: 789,
        acl: TestAcl(),
        defaultAcl: TestAcl(),
    })
}

export interface SpeciesData {
    _id: string
    scientificName: string
    aliases?: string[]
    standardSubstrate: string
    notes?: Note[]
    lastUpdated: number
    acl: ACL
    defaultAcl: ACL
}
export class SpeciesData {
    // Accept a single object containing the fields
    constructor(init?: Partial<SpeciesData>) {
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
        return "species"
    }
}

// TODO: there is an alternative to this, so we may not want this or to use it
export function SpeciesSelectorCloseable(sp: SelectorProps<SpeciesData>) { // TODO: use
    const doSel = (val?: SpeciesData):void=>{
        if (!val){
            return
        }
        sp.doSelect(val)
    }
    return <CloseableSelector<SpeciesData> props={{
        allowCreation: sp.allowCreation,
        doSelect: doSel, // For selecting normally
        closeTxt: "Close Species List",
        //createTxt: "Create Species",// TODO: ???
        lowercase: "species",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "species",// TODO: ???
        createSelector:(selHdl: (onSelect: SpeciesData) => void)=>{
            return <SpeciesSelector doSelect={(v)=>{
                v && selHdl(v)
            }}/>
        },
        // TODO: ok?
        // createCreator:(selHdl: (onSelect: FruitingChamberData) => void)=>{
        //     return <NewFruitingChamberForm handlers={{onCreate: selHdl, isTopLevel: false}}/>
        // },
    }}/>
}
