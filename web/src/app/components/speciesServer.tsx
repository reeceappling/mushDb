import {Note} from "@/app/components/formSubcomponents/notes";
import {EntryPerms} from "@/app/components/perms";
import {ACL} from "@/app/components/accessControlServer";
import CloseableSelector, {SelectorProps} from "@/app/components/selector";
import {ChannelTextNewAgarBatch} from "@/app/components/agarBatchServer";
import {SlantSelector} from "@/app/components/slantClient";
import {SlantData} from "@/app/components/slantServer";
import {SpeciesSelector} from "@/app/components/speciesClient";

export function TestSpeciesOk() {
    const a: SpeciesData = {
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
        //perms: {userPerms: {ids:[{id:"userCollId",val:"userName"}],canWrite:[true]},projectPerms: {ids:["proj1","proj2"],canWrite:[true, false]}, blanketPerms: 1},
    }
    return a
}

export interface SpeciesData {
    _id: string
    scientificName: string
    aliases?: string[]
    standardSubstrate: string
    notes?: Note[]
    lastUpdated: number
    acl?: ACL
    defaultAcl?: ACL
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
        msgTxt: ChannelTextNewAgarBatch, // TODO: ???
        closeTxt: "Close Species List",
        //createTxt: "Create Species",// TODO: ???
        lowercase: "species",
        //creatorInPage: sp.creatorInPage,// TODO: ???
        //createEndpt: "species",// TODO: ???
        getId: (v: SpeciesData) => v._id,
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
